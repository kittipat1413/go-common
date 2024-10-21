package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kittipat1413/go-common/framework/errors"
	"github.com/kittipat1413/go-common/framework/logger"
)

// To run this example with Gin, execute the following commands:
// 1. go run main.go
// 2. curl -X GET http://localhost:8080/users
// 3. curl -X GET http://localhost:8080/custom-error
// 4. curl -X POST http://localhost:8080/login -d '{"username": "admin", "password": "password"}'

func main() {
	// Initialize the error handling framework with the service prefix.
	errors.SetServicePrefix("USER-SVC")

	// Create a Gin router.
	router := gin.Default()

	// Define routes.
	router.GET("/users", GetUsers)
	router.GET("/custom-error", GetCustomError)
	router.POST("/login", Login)

	// Start the server.
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}

// GetUsers handles GET /users requests.
func GetUsers(c *gin.Context) {
	users, err := getUsers()
	if err != nil {
		domainErr, _ := errors.NewNotFoundError()
		ErrorResp(c, domainErr.Wrap(err))
		return
	}

	// Return the list of users.
	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func getUsers() (data []string, err error) {
	errLocation := "[Service getUsers]"
	defer errors.WrapError(errLocation, &err)

	// Simulate Database Error
	return nil, sql.ErrNoRows
}

// GetCustomError handles GET /invalid-error requests.
func GetCustomError(c *gin.Context) {
	err := &InvalidCustomError{}
	ErrorResp(c, err)
}

// Login handles POST /login requests.
func Login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Bind JSON input.
	if err := c.ShouldBindJSON(&credentials); err != nil {
		errResp, newErr := NewMissingUserPasswordError()
		if newErr != nil {
			ErrorResp(c, newErr)
			return
		}
		ErrorResp(c, errResp)
		return
	}

	// Simulate authentication failure.
	if credentials.Username != "admin" || credentials.Password != "password" {
		errResp, _ := errors.NewUnauthorizedError()
		ErrorResp(c, errResp)
		return
	}

	// Return success response if authentication is successful.
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
}

////////////////////////////////////////////////////
// CUSTOM ERROR DEFINITIONS
// This section demonstrates how to define custom errors in your application.
////////////////////////////////////////////////////

const (
	StatusCodeMissingUserPassword = "91001"
)

// MissingUserPasswordError represents an error when username or password is missing.
type MissingUserPasswordError struct {
	*errors.BaseError
}

// NewMissingUserPasswordError creates a new MissingUserPasswordError instance.
func NewMissingUserPasswordError() (*MissingUserPasswordError, error) {
	baseErr, err := errors.NewBaseError(
		StatusCodeMissingUserPassword,
		"Missing username or password",
		http.StatusBadRequest,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &MissingUserPasswordError{
		BaseError: baseErr,
	}, nil
}

// InvalidCustomError is an example error type that does not embed BaseError.
// Because it does not follow the standard error framework, it will not
// be properly handled by the error handling functions, resulting in a
// generic internal server error response.
type InvalidCustomError struct{}

func (e *InvalidCustomError) Code() string {
	return "xyz-123"
}

func (e *InvalidCustomError) GetMessage() string {
	return "some error occurred"
}

func (e *InvalidCustomError) GetHTTPCode() int {
	return http.StatusInternalServerError
}

func (e *InvalidCustomError) Error() string {
	return e.GetMessage()
}

func (e *InvalidCustomError) Wrap(err error) error {
	return fmt.Errorf("%w: %v", e, err)
}

func (e *InvalidCustomError) GetData() interface{} {
	return nil
}

////////////////////////////////////////////////////
// ERROR RESPONSE STRUCTURE AND HANDLING
// This section demonstrates how to structure and handle error responses in your application.
////////////////////////////////////////////////////

// ErrorResponse represents the structure of the error response.
type ErrorResponse struct {
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	HTTPCode int         `json:"-"`
	Data     interface{} `json:"data,omitempty"`
}

// ErrorResp sends an error response with the appropriate status code.
func ErrorResp(c *gin.Context, err error) {
	// Unwrap the error and extract the error response.
	errObj := unwrapError(err)
	// Log the error.
	log := logger.FromContext(c.Request.Context())
	log.Error(c.Request.Context(), errObj.Message, err, nil)
	// Send the error response.
	c.AbortWithStatusJSON(errObj.HTTPCode, errObj)
}

// unwrapError processes the error and extracts information for the response.
// It handles DomainError and standard errors.
func unwrapError(err error) ErrorResponse {
	errResp := ErrorResponse{
		Code:     errors.GetFullCode(errors.StatusCodeGenericInternalError),
		Message:  err.Error(),
		HTTPCode: http.StatusInternalServerError,
	}

	// Try to unwrap the error and find a valid DomainError.
	if domainErr := errors.UnwrapDomainError(err); domainErr != nil {
		errResp.Code = domainErr.Code()
		errResp.Message = domainErr.GetMessage()
		errResp.HTTPCode = domainErr.GetHTTPCode()
		errResp.Data = domainErr.GetData()
	}

	return errResp
}
