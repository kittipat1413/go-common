package errors

import (
	"strings"
)

var servicePrefix = "ERR" // Default service prefix

// SetServicePrefix sets the service-specific prefix (e.g., "USER-SVC"). It converts the prefix to uppercase to maintain consistency.
func SetServicePrefix(prefix string) {
	servicePrefix = strings.ToUpper(prefix)
}

// GetServicePrefix returns the current service prefix.
func GetServicePrefix() string {
	return servicePrefix
}
