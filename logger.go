package logger

import (
	"io"
	"os"

	"context"

	"fmt"

	"cloud.google.com/go/errorreporting"
	"github.com/mgutz/logxi/v1"
	"github.com/pkg/errors"
)

// UnaLogger wraps a logxi logger
// and delegate to some of it's logging methods
type UnaLogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
	Fatal(msg string, err error, args ...interface{})
	Underlying() log.Logger
}

type unaLogger struct {
	Logger log.Logger
	name   string
}

// Config contains Name and FileName for the logger
type Config struct {
	Name     string
	FileName string
}

// Keep a list of loggers that we can use in the SetLevel func
var loggers []UnaLogger

// SetUpErrorReporting creates an ErrorReporting client and returns that client together with a reportPanics function.
// That function should be defered in every new scope where you want to catch pancis and have them pass on to Stackdriver
// Error Reporting
func SetUpErrorReporting(ctx context.Context, projectID, serviceName, serviceVersion string) (client *errorreporting.Client, reportPanics func()) {
	lgr := New("errorreporting")
	errorClient, err := errorreporting.NewClient(ctx, projectID,
		errorreporting.Config{
			ServiceName:    serviceName,
			ServiceVersion: serviceVersion})
	if err != nil {
		lgr.Fatal("Couldn't create an errorreporting client", err)
	}
	return errorClient, func() {
		x := recover()
		if x == nil {
			return
		}
		switch e := x.(type) {
		case string:
			err := errorClient.ReportSync(ctx, errorreporting.Entry{Error: errors.New(e)})
			if err != nil {
				lgr.Error("Couldn't do a ReportSync to Stackdriver Error Reporting", err)
			}
		}
		// repanics so the app execution stops
		panic(fmt.Sprintf("Repanicked from logger: %v", x))
	}
}

// New creates a new logger with the given (string) name
func New(name string) UnaLogger {
	return NewLogger(Config{Name: name})
}

// NewLogger creates a new logger with the given (Config) name
func NewLogger(conf Config) UnaLogger {
	// These configurations are made to make the
	// log payload compatible with the LogEntry format used in Google Cloud Logging
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
	log.KeyMap.Level = "severity"
	log.KeyMap.Message = "message"
	log.KeyMap.Time = "timestamp"
	log.LevelMap[log.LevelError] = "ERROR"
	log.LevelMap[log.LevelInfo] = "INFO"
	log.LevelMap[log.LevelDebug] = "DEBUG"

	logxiLogger := log.New(conf.Name)
	if conf.FileName != "" {
		if file, err := os.Create(conf.FileName); err == nil {
			logxiLogger = log.NewLogger(file, conf.Name)
		}
	}

	return &unaLogger{
		Logger: logxiLogger,
		name:   conf.Name,
	}
}

// SetLevel loops over loggers and sets the level on each logger
func SetLevel(level int) {
	for _, logger := range loggers {
		logger.Underlying().SetLevel(level)
	}
}

// SetWriter overrides the io.Writer of the underlying logxi logger
func (ul *unaLogger) SetWriter(writer io.Writer) {
	ul.Logger = log.NewLogger(writer, ul.name)
}

// Underlying returns the underlying logxi logger
func (ul unaLogger) Underlying() log.Logger {
	return ul.Logger
}

// Info logs to Stdout with an "INFO" prefix
func (ul unaLogger) Info(msg string, args ...interface{}) {
	ul.Logger.Info(msg, args...)
}

// Debug logs to Stdout with an "DEBUG" prefix if Debug level is enabled
func (ul unaLogger) Debug(msg string, args ...interface{}) {
	if ul.Logger.IsDebug() {
		ul.Logger.Debug(msg, args...)
	}
}

// Error logs to Stdout with an "Error" prefix
// It also adds an "error" key to the provided err(error) argument
func (ul unaLogger) Error(msg string, err error, args ...interface{}) {
	_ = ul.Logger.Error(msg, append(args, "error", err)...)
}

// Fatal logs to Stdout with an "Fatal" prefix
// It also adds an "error" key to the provided err(error) argument
func (ul unaLogger) Fatal(msg string, err error, args ...interface{}) {
	ul.Logger.Fatal(msg, append(args, "error", err)...)
}
