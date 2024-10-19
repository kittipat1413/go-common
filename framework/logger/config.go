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
	if level, ok := logrusLevelMapper[l]; ok {
		return level
	}
	// Default to InfoLevel if unknown
	return logrus.InfoLevel
}

func (l LogLevel) IsValid() bool {
	_, ok := logrusLevelMapper[l]
	return ok
}

const (
	// DefaultEnvironmentKey is the default key used for the environment field in logs.
	DefaultEnvironmentKey = "environment"
	// DefaultServiceNameKey is the default key used for the service name field in logs.
	DefaultServiceNameKey = "service_name"
	// DefaultErrorKey is the default key used for the error field in logs.
	DefaultErrorKey = "error"
)
