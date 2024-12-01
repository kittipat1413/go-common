package errors

import (
	"strings"
)

const DefaultServicePrefix = "ERR" // DefaultServicePrefix is the default prefix used for errors.

var (
	servicePrefix = DefaultServicePrefix
)

// SetServicePrefix sets the service-specific prefix (e.g., "USER-SVC"). It converts the prefix to uppercase to maintain consistency.
// If an empty prefix is provided, the default prefix (ERR) is used.
func SetServicePrefix(prefix string) {
	if prefix == "" {
		servicePrefix = DefaultServicePrefix
	} else {
		servicePrefix = strings.ToUpper(prefix)
	}

}

// GetServicePrefix returns the current service prefix.
func GetServicePrefix() string {
	return servicePrefix
}
