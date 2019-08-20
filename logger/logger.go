package logging

import (
	"os"

	logrus "github.com/sirupsen/logrus"
	// prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"path"
	"runtime"
	"strings"
)

// DefaultLogLevel sets the default log level
var DefaultLogLevel logrus.Level

func init() {
	DefaultLogLevel = logrus.ErrorLevel
}

// Fields allows passing key value pairs to Logrus
type Fields map[string]interface{}

// ContextHook provides an interface for Logrus Hook for callbacks
type ContextHook struct{}

// Levels returns all available logging levels, required for Logrus Hook implementation
func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is a callback issued for every log message, allowing us to modify the output, required for Logrus Hook implementation
func (hook ContextHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 5, 5)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 2)
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") &&
			!strings.Contains(name, "github.com/coccyx/gogen/logger") {
			file, line := fu.FileLine(pc[i] - 2)
			entry.Data["file"] = path.Base(file)
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			break
		}
	}
	return nil
}

func init() {
	// logrus.SetFormatter(&prefixed.TextFormatter{TimestampFormat: "Jan 02 15:04:05.000"})
	logrus.AddHook(ContextHook{})
	logrus.SetLevel(DefaultLogLevel)
}

// WithField adds a field to the logrus entry
func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}

// WithFields add fields to the logrus entry
func WithFields(fields Fields) *logrus.Entry {
	sendfields := make(logrus.Fields)
	for k, v := range fields {
		sendfields[k] = v
	}
	return logrus.WithFields(sendfields)
}

// WithError adds an error field to the logrus entry
func WithError(err error) *logrus.Entry {
	return logrus.WithError(err)
}

// Debugf logs a message at Debug level
func Debugf(format string, v ...interface{}) {
	logrus.Debugf(format, v...)
}

// Infof logs a message at Info level
func Infof(format string, v ...interface{}) {
	logrus.Infof(format, v...)
}

// Warningf logs a message at Warning level
func Warningf(format string, v ...interface{}) {
	logrus.Warningf(format, v...)
}

// Errorf logs a message at Error level
func Errorf(format string, v ...interface{}) {
	logrus.Errorf(format, v...)
}

// Error logs a message at Error level
func Error(v ...interface{}) {
	logrus.Error(v...)
}

// Warning logs a message at Warning level
func Warning(v ...interface{}) {
	logrus.Warning(v...)
}

// Info logs a message at Info level
func Info(v ...interface{}) {
	logrus.Info(v...)
}

// Debug logs a message at Debug level
func Debug(v ...interface{}) {
	logrus.Debug(v...)
}

// Panic logs a message at Panic and then exits
func Panic(v ...interface{}) {
	logrus.Panic(v...)
}

// Panicf logs a message at Panic and then exits
func Panicf(format string, v ...interface{}) {
	logrus.Panicf(format, v...)
}

// Fatal logs a message and then exits
func Fatal(v ...interface{}) {
	logrus.Fatal(v...)
}

// Fatalf logs a message and then exits
func Fatalf(format string, v ...interface{}) {
	logrus.Fatalf(format, v...)
}

// EnableJSONOutput sets the Logrus formatter to output in JSON
func EnableJSONOutput() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.AddHook(ContextHook{})
}

// EnableTextOutput sets the Logrus formatter to output as plain text
func EnableTextOutput() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}

// SetOutput sets output a specified file name
func SetOutput(name string) {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, DisableColors: true})
	out, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		logrus.SetOutput(os.Stderr)
	}
	logrus.SetOutput(out)
}

// SetDebug the log level to Debug
func SetDebug(on bool) {
	if on {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.AddHook(ContextHook{})
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}
}

// SetInfo sets the log level to Info
func SetInfo() {
	logrus.SetLevel(logrus.InfoLevel)
}

// SetWarn sets the log level to Warning
func SetWarn() {
	logrus.SetLevel(logrus.WarnLevel)

}
