package errors

import (
	"fmt"
)

func WrapError(prefix string, errptr *error) {
	if *errptr != nil {
		*errptr = fmt.Errorf(prefix+": %w", *errptr)
	}
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
