package logger

import (
	"io"
	"os"

	"context"

	"fmt"

	"sync"

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

var errorClient *errorreporting.Client

// logger for internal use
var lgr UnaLogger

// InitErrorReporting will enable the errors of all calls to Error and Fatal to be sent to Google Error Reporting
// It also enables the
func InitErrorReporting(ctx context.Context, projectID, serviceName, serviceVersion string) error {
	lgr = New("unalogger")
	client, err := errorreporting.NewClient(ctx, projectID,
		errorreporting.Config{
			ServiceName:    serviceName,
			ServiceVersion: serviceVersion})
	if err != nil {
		return err
	}

	errorClient = client

	return nil
}

// ReportPanics should be defered in every new scope where you want to catch pancis and have them pass on to Stackdriver
// Error Reporting
func ReportPanics(ctx context.Context) func() {
	return func() {
		if errorClient == nil {
			panic("The errorClient was nil, initialize it with InitErrorReporting before deferring this function")
		}
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
		panic(fmt.Sprintf("Repanicked from logger: %s", x))
	}
}

// CloseClient should be deferred right after calling InitErrorReporting to enure that the client is
// closed down gracefully
func CloseClient() {
	if errorClient == nil {
		panic("The errorClient was nil, initialize it with InitErrorReporting before deferring this function")
	}
	var _ = errorClient.Close() // Ignoring this error
}

// Deprecated: The functionality is split into InitErrorReporting, ReportPanics and CloseClient instead
// SetUpErrorReporting creates an ErrorReporting client and returns that client together with a reportPanics function.
// That function should be defered in every new scope where you want to catch pancis and have them pass on to Stackdriver
// Error Reporting
func SetUpErrorReporting(ctx context.Context, projectID, serviceName, serviceVersion string) (client *errorreporting.Client, reportPanics func()) {
	lgr := New("errorreporting")
	errClient := InitErrorReporting(ctx, projectID, serviceName, serviceVersion)
	if errClient != nil {
		lgr.Fatal("Couldn't create an errorreporting client", errClient)
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

func reportError(err error) {
	if errorClient != nil {
		errorClient.Report(errorreporting.Entry{
			Error: err,
		})
	}
}

var defaultsSet bool
var mutex = sync.Mutex{}

func setDefaults() {
	// These configurations are made to make the
	// log payload compatible with the LogEntry format used in Google Cloud Logging
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
	log.KeyMap.Level = "severity"
	log.KeyMap.Message = "message"
	log.KeyMap.Time = "timestamp"
	log.LevelMap[log.LevelFatal] = "CRITICAL"
	log.LevelMap[log.LevelError] = "ERROR"
	log.LevelMap[log.LevelInfo] = "INFO"
	log.LevelMap[log.LevelDebug] = "DEBUG"
	defaultsSet = true
}

// New creates a new logger with the given (string) name
func New(name string) UnaLogger {
	return NewLogger(Config{Name: name})
}

// NewLogger creates a new logger with the given (Config) name
func NewLogger(conf Config) UnaLogger {

	logxiLogger := log.New(conf.Name)
	if conf.FileName != "" {
		if file, err := os.Create(conf.FileName); err == nil {
			logxiLogger = log.NewLogger(file, conf.Name)
		}
	}
	unaLogger := &unaLogger{
		Logger: logxiLogger,
		name:   conf.Name,
	}

	// Add the logger to the list of loggers and set some defaults
	// Needs to use a mutex here so loggers can be created in different goroutines
	mutex.Lock()
	loggers = append(loggers, unaLogger)
	if !defaultsSet {
		setDefaults()
	}
	mutex.Unlock()

	return unaLogger
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
	reportError(err)
	_ = ul.Logger.Error(msg, append(args, "error", err)...)
}

// Fatal logs to Stdout with an "Fatal" prefix
// It also adds an "error" key to the provided err(error) argument
func (ul unaLogger) Fatal(msg string, err error, args ...interface{}) {
	reportError(err)
	ul.Logger.Fatal(msg, append(args, "error", err)...)
}
