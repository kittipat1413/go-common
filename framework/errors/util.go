package errors

import (
	"fmt"
)

// WrapErrorWithPrefix wraps the input error with a prefix. If the error is nil, it does nothing.
func WrapErrorWithPrefix(prefix string, errptr *error) {
	if *errptr != nil {
		*errptr = fmt.Errorf(prefix+": %w", *errptr)
	}
}

// WrapError wraps two errors into one. If either error is nil, it returns the non-nil error. If both are non-nil, it wraps the new error around the original error.
func WrapError(original, new error) error {
	if original == nil && new == nil {
		return nil
	}
	if original == nil {
		return new
	}
	if new == nil {
		return original
	}
	// Wrap the new error around the original error
	return fmt.Errorf("%w: %v", new, original)
}

// UnwrapDomainError attempts to find a DomainError in the error chain. The error should implement the DomainError interface and have a BaseError embedded.
// It unwraps the error chain and checks each error to see if it is a DomainError and if it contains a BaseError. If such an error is found, it is returned.
func UnwrapDomainError(err error) DomainError {
	unwrapErr := err
	for unwrapErr != nil {
		// Check if the error explicitly implements DomainError and has a BaseError.
		if domainErr, ok := unwrapErr.(DomainError); ok && ExtractBaseError(domainErr) != nil {
			return domainErr
		}

		// Try to unwrap the next error in the chain.
		type unwrapper interface {
			Unwrap() error
		}
		// If the error does not implement an unwrapper, stop unwrapping.
		if unwrappableErr, ok := unwrapErr.(unwrapper); ok {
			unwrapErr = unwrappableErr.Unwrap()
		} else {
			break
		}
	}
	return nil
}
