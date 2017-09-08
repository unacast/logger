package logger

import (
	"fmt"

	"github.com/getsentry/raven-go"
	"github.com/mgutz/logxi/v1"
	"io"
	"os"
)

// UnaLogger wraps a logxi logger
// and delegate to some of it's logging methods
type UnaLogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
	Underlying() log.Logger
	PassToSentry()
}

type unaLogger struct {
	Logger log.Logger
	passToSentry bool
	name string
}

type Config struct {
	Name string
	FileName string
}

// NewLogger creates a new logger with the given name
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
		name: conf.Name,
	}
}

// NewLogger creates a new logger with the given name
// and that passes errors to Sentry
func NewSentryLogger(conf Config) UnaLogger {
	l := NewLogger(conf)
	l.PassToSentry()
	return l
}

// SetWriter overrides the io.Writer of the underlying logxi logger
func (ul *unaLogger) SetWriter(writer io.Writer)  {
	ul.Logger = log.NewLogger(writer, ul.name)
}

// PassToSentry indicates whether the Error function
// should pass errors on to Sentry or not
func (ul *unaLogger) PassToSentry() {
	ul.passToSentry = true
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
// If
func (ul unaLogger) Error(msg string, err error, args ...interface{}) {
	tags := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		tags[args[i].(string)] = fmt.Sprintf("%v", args[i+1])
	}
	e := ul.Logger.Error(msg, "error", err, "labels", tags)

	if ul.passToSentry {
		raven.SetDefaultLoggerName(ul.name)
		raven.CaptureError(e, tags)
	}
}
