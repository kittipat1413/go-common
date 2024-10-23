# Errors Package
The errors package provides a centralized error handling library for Go applications. It standardizes error codes, messages, and HTTP status codes across services, making error management consistent and efficient.

## Features
- **Standardized Error Codes**: Uses a consistent error code format (`xyzzz`) across services.
- **Service Prefixes**: Allows setting a service-specific prefix for error codes.
- **Domain Errors**: Provides a `DomainError` interface for custom errors.
- **Base Error Embedding**: Encourages embedding `BaseError` for consistency.
- **Utilities**: Includes helper functions for wrapping, unwrapping, and extracting errors.
- **Category Validation**: Validates that error codes align with predefined categories.

## Getting Started

### Setting the Service Prefix
Before using the error handling library, set the service-specific prefix. This helps in identifying which service an error originated from.
```golang
import "github.com/kittipat1413/go-common/framework/errors"

func init() {
    errors.SetServicePrefix("USER-SVC")
}
```

### Defining Custom Errors
To define a custom error, create a new type that embeds `*errors.BaseError` and implements the `DomainError` interface.
```golang
package myerrors

import (
    "fmt"
    "net/http"

    "github.com/kittipat1413/go-common/framework/errors"
)

const (
    StatusCodeUserNotFound = "23001"
)

type UserNotFoundError struct {
    *errors.BaseError
}

func NewUserNotFoundError(userID string) (*UserNotFoundError, error) {
    baseErr, err := errors.NewBaseError(
        StatusCodeUserNotFound,
        fmt.Sprintf("User with ID %s not found", userID),
        http.StatusNotFound,
        map[string]string{"user_id": userID},
    )
    if err != nil {
        return nil, err
    }
    return &UserNotFoundError{BaseError: baseErr}, nil
}
```

### Using the Error Handling Utilities

**Adding Context with a Prefix**: Use `errors.WrapErrorWithPrefix` to add context to an error with a specified prefix. This helps in tracking where the error occurred. If the error is nil, it does nothing.
```golang
func someFunction() (err error) {
    defer errors.WrapErrorWithPrefix("[someFunction]", &err)
    // Function logic...
    return
}
```
**Wrapping Errors**: Use `errors.WrapError` to combine multiple errors into one. If either error is nil, it returns the non-nil error. If both are non-nil, it wraps the new error around the original error.
```golang
user, err := getUser()
if err != nil {
    // Creating a domain-specific error
    domainErr := errors.New("user not found")

    // Wrapping the domain error with the original error
    return errors.WrapError(err, domainErr)
}
```
**Unwrapping Domain Errors**: Use `errors.UnwrapDomainError` to extract the `DomainError` from an error chain, allowing for specialized handling of domain-specific errors.
```golang
func handleError(err error) {
    if domainErr := errors.UnwrapDomainError(err); domainErr != nil {
        // Handle domain error
    } else {
        // Handle generic error
    }
}
```

## Error Code Convention
Error codes follow the `xyzzz` format:
- `x`: Main category (e.g., 2 for Client Errors).
- `y`: Subcategory (e.g., 1 for Invalid Parameters).
- `zzz`: Specific error code (e.g., 001 for a particular invalid parameter).
> **Example**: `21001` could represent an invalid username parameter.

## Error Categories
Defined categories and their descriptions:
- `10`: Success
- `20`: Client Errors
    - `21`: Invalid Parameters
    - `22`: Duplicated Entry
    - `23`: Not Found
    - `24`: Unprocessable Entity
- `50`: Internal Errors
    - `51`: Database Errors
    - `52`: Third-party Errors
- `90`: Security Errors
    - `91`: Unauthorized
    - `92`: Forbidden
> The validCategories map in `categories.go` maintains the valid categories and their descriptions.

## Examples
You can find a complete working example in the repository under [framework/errors/example](example/).

## Best Practices
- **Consistency**: Always define error codes and messages in a centralized place to maintain consistency.
- **Embedding BaseError**: Ensure all custom errors embed `*errors.BaseError` to integrate with the error handling utilities.
- **Category Alignment**: When defining new error codes, make sure they align with the predefined categories and use the correct HTTP status codes.
- **Avoid Manual Synchronization**: Use the centralized data structures for categories and error codes to prevent inconsistencies.