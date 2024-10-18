package logger

import "github.com/sirupsen/logrus"

type LogLevel string

const (
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
	FATAL LogLevel = "fatal"
)

var logrusLevelMapper = map[LogLevel]logrus.Level{
	DEBUG: logrus.DebugLevel,
	INFO:  logrus.InfoLevel,
	WARN:  logrus.WarnLevel,
	ERROR: logrus.ErrorLevel,
	FATAL: logrus.FatalLevel,
}

func (l LogLevel) ToLogrusLevel() logrus.Level {
	return logrusLevelMapper[l]
}

const (
	// DefaultEnvironmentKey is the default key used for the environment field in logs.
	DefaultEnvironmentKey = "environment"
	// DefaultServiceNameKey is the default key used for the service name field in logs.
	DefaultServiceNameKey = "service_name"
	// DefaultErrorKey is the default key used for the error field in logs.
	DefaultErrorKey = "error"
)

// FieldKeyFormatter is a function type that allows users to customize log field keys.
type FieldKeyFormatter func(key string) string

// NoopFieldKeyFormatter is the default implementation of FieldKeyFormatter,
// which returns the key unchanged.
func NoopFieldKeyFormatter(defaultKey string) string {
	return defaultKey
}
