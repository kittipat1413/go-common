package logger

import "github.com/sirupsen/logrus"

// LogLevel represents the logging level as a string type for easy configuration.
type LogLevel string

// Log level constants define the available logging levels.
// These levels determine which messages are logged based on severity.
const (
	DEBUG LogLevel = "debug" // Verbose output for debugging (lowest priority)
	INFO  LogLevel = "info"  // General informational messages
	WARN  LogLevel = "warn"  // Warning messages for potentially harmful situations
	ERROR LogLevel = "error" // Error messages for failure conditions
	FATAL LogLevel = "fatal" // Critical errors that may cause application termination
)

// logrusLevelMapper provides mapping from custom LogLevel to logrus levels.
// Enables conversion between string-based levels and logrus internal representation.
var logrusLevelMapper = map[LogLevel]logrus.Level{
	DEBUG: logrus.DebugLevel,
	INFO:  logrus.InfoLevel,
	WARN:  logrus.WarnLevel,
	ERROR: logrus.ErrorLevel,
	FATAL: logrus.FatalLevel,
}

// ToLogrusLevel converts the custom LogLevel to the corresponding logrus.Level.
// Returns logrus.InfoLevel as a safe default for unknown/invalid levels.
func (l LogLevel) ToLogrusLevel() logrus.Level {
	if level, ok := logrusLevelMapper[l]; ok {
		return level
	}
	// Default to InfoLevel if unknown
	return logrus.InfoLevel
}

// IsValid checks if the LogLevel is one of the defined valid levels.
// Useful for configuration validation and input sanitization.
func (l LogLevel) IsValid() bool {
	_, ok := logrusLevelMapper[l]
	return ok
}

// Default field key constants for structured logging.
// These keys ensure consistent field naming across all log entries
// and enable proper log parsing and filtering in log aggregation systems.
const (
	DefaultEnvironmentKey = "environment"  // Default key for the environment field in logs (e.g., production, staging)
	DefaultServiceNameKey = "service_name" // Default key for the service name field in logs
	DefaultErrorKey       = "error"        // Default key for the error field in logs
)
