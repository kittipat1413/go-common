package errors

import (
	"fmt"
)

// WrapErrorWithPrefix wraps an error pointed to by errptr with a descriptive prefix.
// This is a convenience function for adding context to errors using a pointer pattern,
// which is useful in functions that modify error values in-place.
//
// Example:
//
//	func ProcessFile(filename string) (err error) {
//		defer func() {
//			// Add context to any error that occurred
//			errors.WrapErrorWithPrefix("failed to process file "+filename, &err)
//		}()
//
//		// ... rest of processing
//		return nil
//	}
func WrapErrorWithPrefix(prefix string, errptr *error) {
	if *errptr != nil {
		*errptr = fmt.Errorf(prefix+": %w", *errptr)
	}
}

// WrapError combines two errors into a single error with proper error chaining.
// This function handles various nil combinations gracefully and uses Go's error
// wrapping semantics to preserve the error chain for unwrapping operations.
//
// Error Combination Logic:
//   - If both errors are nil: returns nil
//   - If only one error is nil: returns the non-nil error
//   - If both errors are non-nil: wraps new around original using fmt.Errorf with %w verb
//
// The wrapping order places 'new' as the outer error and 'original' as the inner error,
// allowing proper unwrapping via errors.Unwrap(), errors.Is(), and errors.As().
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

// UnwrapDomainError traverses an error chain to find the first error that implements
// the DomainError interface and contains an embedded BaseError. This function is
// essential for extracting domain-specific error information from wrapped error chains.
//
// Search Algorithm:
//  1. Start with the provided error
//  2. Check if error implements DomainError interface
//  3. Verify that the error contains an embedded BaseError (via ExtractBaseError)
//  4. If both conditions are met, return the domain error
//  5. Otherwise, unwrap to the next error in the chain
//  6. Repeat until chain is exhausted
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
