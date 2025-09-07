package errors

import (
	"strings"
)

// DefaultServicePrefix is the default prefix used for error codes when no custom prefix is configured.
// This prefix is automatically prepended to all error codes to create fully qualified error identifiers.
//
// The default prefix "ERR" results in error codes like "ERR-400001", "ERR-500002", etc.
// This can be customized using SetServicePrefix() to provide service-specific prefixes
// like "USER-SVC-400001" or "AUTH-400001".
const DefaultServicePrefix = "ERR"

var (
	// servicePrefix holds the current service-specific prefix used for error codes.
	// This variable is modified by SetServicePrefix() and accessed by GetServicePrefix().
	// It's automatically converted to uppercase to maintain consistency across the system.
	servicePrefix = DefaultServicePrefix
)

// SetServicePrefix configures a custom service-specific prefix for all error codes.
// The prefix is automatically converted to uppercase to ensure consistency across
// the application. If an empty string is provided, the system reverts to the default prefix.
//
// This function is typically called once during application initialization to establish
// service-specific error code formatting. The prefix helps identify which service
// generated an error in distributed systems or microservice architectures.
//
// Prefix Format:
//   - Should be descriptive of the service (e.g., "USER-SVC", "AUTH", "PAYMENT")
//   - Automatically converted to uppercase for consistency
//   - Should not include the trailing dash separator (automatically added)
//   - Empty strings revert to DefaultServicePrefix
func SetServicePrefix(prefix string) {
	if prefix == "" {
		servicePrefix = DefaultServicePrefix
	} else {
		servicePrefix = strings.ToUpper(prefix)
	}

}

// GetServicePrefix returns the currently configured service prefix.
// This function is primarily used internally by the error formatting system
// but can also be used by applications that need to know the current prefix
// for logging or debugging purposes.
//
// Returns:
//   - string: The current service prefix in uppercase format
func GetServicePrefix() string {
	return servicePrefix
}
