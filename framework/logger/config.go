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

// Fields is a map of fields to add to a log entry.
const (
	environmentKey = "environment"
	serviceNameKey = "service_name"
)
