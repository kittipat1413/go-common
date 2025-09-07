[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Errors Package
The errors package provides a centralized error handling library for Go applications. It standardizes error codes, messages, and HTTP status codes across services, making error management consistent and efficient.

## Features
- **Standardized Error Codes**: Uses a consistent error code format (`xyyzzz`) across services.
- **Service Prefixes**: Allows setting a service-specific prefix for error codes.
- **Domain Errors**: Provides a `DomainError` interface for custom errors.
- **Base Error Embedding**: Encourages embedding `BaseError` for consistency.
- **Utilities**: Includes helper functions for wrapping, unwrapping, and extracting errors.
- **Category Validation**: Validates that error codes align with predefined categories.

## Installation
```bash
go get github.com/kittipat1413/go-common/framework/errors
```

## Documentation
[![Go Reference](https://pkg.go.dev/badge/github.com/kittipat1413/go-common/framework/errors.svg)](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/errors)

For detailed API documentation, examples, and usage patterns, visit the [Go Package Documentation](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/errors).

## Getting Started

### Setting the Service Prefix
Before using the error handling library, set the service-specific prefix. This helps in identifying which service an error originated from.
```go
import "github.com/kittipat1413/go-common/framework/errors"

func init() {
    errors.SetServicePrefix("USER-SVC")
}
```

### Defining Custom Errors
Define custom errors by embedding `*errors.BaseError` in your error type. This ensures that custom errors conform to the `DomainError` interface and can be properly handled by the error utilities.
```go
package myerrors

import (
    "fmt"
    "net/http"

    "github.com/kittipat1413/go-common/framework/errors"
)

const (
    // StatusCodeUserNotFound indicates that the requested user could not be found.
    // Error Code Convention:
    // - Main Category: 4 (Client Errors)
    // - Subcategory: 02 (Not Found)
    // - Specific Error Code: 001 (Defined by the service)
    StatusCodeUserNotFound = "402001"
)

type UserNotFoundError struct {
    *errors.BaseError
}

func NewUserNotFoundError(userID string) (*UserNotFoundError, error) {
    baseErr, err := errors.NewBaseError(
        StatusCodeUserNotFound,
        fmt.Sprintf("User with ID %s not found", userID),
        map[string]string{"user_id": userID},
    )
    if err != nil {
        return nil, err
    }
    return &UserNotFoundError{BaseError: baseErr}, nil
}
```
### Simplify Error Constructors
To avoid handling errors every time you create a custom error, you can design your constructors to handle any internal errors themselves. This way, your error creation functions can have a simpler signature, 
returning only the custom error. You can handle internal errors by:

- **Panicking** 

    If `errors.NewBaseError` returns an `error`, it likely indicates a misconfiguration or coding error (e.g., invalid error code). In such cases, it's acceptable to panic during development to catch the issue early.
    ```go
    func NewUserNotFoundError(userID string) *UserNotFoundError {
        baseErr, err := errors.NewBaseError(
            StatusCodeUserNotFound,
            fmt.Sprintf("User with ID %s not found", userID),
            map[string]string{"user_id": userID},
        )
        if err != nil {
            panic(fmt.Sprintf("Failed to create BaseError: %v", err))
        }
        return &UserNotFoundError{BaseError: baseErr}
    }
    ```
- **Returning an Error Interface ✅**

    If you want the option to handle the error in the calling function, you can modify your constructor to return an `error` interface.
    This allows proper handling of `ErrBaseErrorCreationFailed`, which is returned when NewBaseError fails due to invalid error codes or categories.
    ```go
    func NewUserNotFoundError(userID string) error {
        baseErr, err := errors.NewBaseError(
            StatusCodeUserNotFound,
            fmt.Sprintf("User with ID %s not found", userID),
            map[string]string{"user_id": userID},
        )
        if err != nil {
            return err
        }
        return &UserNotFoundError{BaseError: baseErr}
    }
    ```
- **Using `init()` to Initialize Predefined Errors**

    You can simplify handling predefined errors by initializing them at the package level. This approach removes the need to handle errors every time you use these predefined errors. If `NewBaseError` fails during initialization (e.g., due to a misconfiguration), `log.Fatal` will immediately halt the program and output the error. This way, issues are caught early at startup rather than during runtime.
    ```go
    package myerrors

    import (
        "fmt"
        "log"
        "net/http"
        "github.com/kittipat1413/go-common/framework/errors"
    )

    // Predefined errors as package-level variables.
    var (
        ErrBadRequest        *BadRequestError
        ErrNotFound          *NotFoundError
    )

    // init initializes predefined errors at package load.
    func init() {
        var err error

        // Initialize BadRequestError
        ErrBadRequest, err = newBadRequestError()
        if err != nil {
            log.Fatal(fmt.Sprintf("failed to initialize ErrBadRequest: %v", err))
        }

        // Initialize NotFoundError
        ErrNotFound, err = newNotFoundError()
        if err != nil {
            log.Fatal(fmt.Sprintf("failed to initialize ErrNotFound: %v", err))
        }
    }

    // BadRequestError is a predefined error for bad request cases.
    type BadRequestError struct {
        *BaseError
    }

    // Helper function to initialize BadRequestError.
    func newBadRequestError() (*BadRequestError, error) {
        baseErr, err := errors.NewBaseError(
            StatusCodeGenericBadRequestError,
            "", // Empty message to use the default message.
            nil,
        )
        if err != nil {
            return nil, err
        }
        return &BadRequestError{BaseError: baseErr}, nil
    }

    // NotFoundError is a predefined error for not found cases.
    type NotFoundError struct {
        *BaseError
    }

    // Helper function to initialize NotFoundError.
    func newNotFoundError() (*NotFoundError, error) {
        baseErr, err := errors.NewBaseError(
            StatusCodeGenericNotFoundError,
            "", // Empty message to use the default message.
            nil,
        )
        if err != nil {
            return nil, err
        }
        return &NotFoundError{BaseError: baseErr}, nil
    }
    ```
    > After defining these errors, you can use them directly in your code by referencing `myerrors.ErrBadRequest` or `myerrors.ErrNotFound`. Since they are pre-initialized at the package level, they are always available without needing additional error handling for creation.

### Why Wrap `BaseError` in a Custom Type?
In Go, it’s common to wrap a base error type inside a more specific domain error type (like `UserNotFoundError`). Here’s why this approach is beneficial:
- **Stronger Type Assertions**
    - When handling errors, using a custom error type allows for better type checking with `errors.As()`.
  ```go
    err := NewUserNotFoundError("12345")

    var userErr *UserNotFoundError
    if errors.As(err, &userErr) {
        fmt.Println("Handling UserNotFoundError:", userErr.GetMessage())
    }
  ```
- **Better Encapsulation of Business Logic**
    - A custom error type keeps domain logic inside the error itself, making it easier to manage.
    ```go
    type UserNotFoundError struct {
        *errors.BaseError
    }

    // Example: Define custom logic for this error
    func (e *UserNotFoundError) IsCritical() bool {
        return false // A missing user is not considered a critical failure
    }
    ```
    ```go
    // Now, error handling can adapt based on business logic:
    var userErr *UserNotFoundError
    if errors.As(err, &userErr) && userErr.IsCritical() {
        // Handle it differently if it's a critical issue
    }
    ```
- **Improves Readability & Maintains Domain Clarity**
    - Without a Custom Error Type:
    ```go
    return errors.NewBaseError(StatusCodeUserNotFound, "User not found", nil)
    ```
    - With a Custom Error Type: 
    ```go
    return NewUserNotFoundError(userID)
    ```

### Using the Error Handling Utilities

**Adding Context with a Prefix**: Use `errors.WrapErrorWithPrefix` to add context to an error with a specified prefix. This helps in tracking where the error occurred. If the error is nil, it does nothing.
```go
func someFunction() (err error) {
    defer errors.WrapErrorWithPrefix("[someFunction]", &err)
    // Function logic...
    return
}
```
**Wrapping Errors**: Use `errors.WrapError` to combine multiple errors into one. If either error is nil, it returns the non-nil error. If both are non-nil, it wraps the new error around the original error.
```go
user, err := getUser()
if err != nil {
    // Creating a domain-specific error
    domainErr := errors.New("user not found")

    // Wrapping the domain error around the original error
    return errors.WrapError(err, domainErr)
}
```
**Unwrapping Domain Errors**: Use `errors.UnwrapDomainError` to extract the `DomainError` from an error chain, allowing for specialized handling of domain-specific errors.
```go
func handleError(err error) {
    if domainErr := errors.UnwrapDomainError(err); domainErr != nil {
        // Handle domain error
    } else {
        // Handle generic error
    }
}
```

## Error Code Convention
Error codes follow the `xyyzzz` format:
- `x`: Main category (e.g., 4 for Client Errors).
- `yy`: Subcategory (e.g., 01 for Bad Request Errors).
- `zzz`: Specific error code (e.g., 001 for a particular invalid parameter).
> **Example**: `401001` could represent an invalid username parameter.

## Error Categories
Defined categories and their descriptions:
- `200zzz`: Success
    - `201zzz`: Partial Success
    - `202zzz`: Accepted
- `400zzz`: Client Errors
    - `401zzz`: Bad Request
    - `402zzz`: Not Found
    - `403zzz`: Conflict
    - `404zzz`: Unprocessable Entity
- `500zzz`: Server Errors
    - `501zzz`: Database Errors
    - `502zzz`: 3rd Party Errors
    - `503zzz`: Service Unavailable
- `900zzz`: Security Errors
    - `901zzz`: Unauthorized
    - `902zzz`: Forbidden
> The validCategories map in `categories.go` maintains the valid categories and their descriptions.

## Examples
You can find a complete working example in the repository under [framework/errors/example](example/).

## Real-World Examples
**[Ticket Reservation System](https://github.com/kittipat1413/ticket-reservation)** - A complete microservice demonstrating error handling patterns, HTTP middleware integration, and domain-specific error types.

## Best Practices
- **Consistency**: Always define error codes and messages in a centralized place to maintain consistency.
- **Embedding BaseError**: Ensure all custom errors embed `*errors.BaseError` to integrate with the error handling utilities.
- **Category Alignment**: When defining new error codes, make sure they align with the predefined categories and use the correct HTTP status codes.
- **Avoid Manual Synchronization**: Use the centralized data structures for categories and error codes to prevent inconsistencies.