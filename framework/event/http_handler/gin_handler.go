package httphandler

import (
	"net/http"

	"github.com/kittipat1413/go-common/framework/event"

	"github.com/gin-gonic/gin"
)

// GinEventHandlerFunc defines a function that processes business logic for an event message.
// This function receives the Gin context and typed event message, returning an error
// if business processing fails.
//
// Parameters:
//   - ctx: Gin context with HTTP request/response and framework utilities
//   - msg: Typed event message containing payload and metadata
//
// Returns:
//   - error: Business logic error (nil indicates success)
//
// Example:
//
//	func processUserEvent(ctx *gin.Context, msg event.EventMessage[UserData]) error {
//	    userData := msg.GetPayload()
//	    return userService.CreateUser(ctx.Request.Context(), userData)
//	}
type GinEventHandlerFunc[T any] func(ctx *gin.Context, msg event.EventMessage[T]) error

// NewGinEventHandler creates a Gin HTTP handler that processes incoming JSON requests as events.
// It integrates event processing lifecycle with HTTP request handling, providing automatic
// JSON unmarshaling, event handler hooks, and standardized HTTP responses.
//
// Request Processing Flow:
//  1. Read and validate JSON request body
//  2. Unmarshal JSON into typed event message
//  3. Execute BeforeHandle hook for pre-processing
//  4. Run business logic via ginHandler function
//  5. Execute AfterHandle hook for post-processing
//  6. Return appropriate HTTP response (200 OK or error status)
//
// Parameters:
//   - ginHandler: Business logic function that processes the event
//   - eventHandler: Event handler providing unmarshaling and lifecycle hooks
//
// Returns:
//   - gin.HandlerFunc: HTTP handler ready for Gin route registration
func NewGinEventHandler[T any](ginHandler GinEventHandlerFunc[T], eventHandler event.EventHandler[T]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Read the raw JSON data
		data, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Unmarshal the raw JSON data into an event message
		msg, err := eventHandler.UnmarshalEventMessage(data)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Pre-processing
		if err := eventHandler.BeforeHandle(ctx.Request.Context(), msg); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Execute the business logic with the event message
		eventResult := ginHandler(ctx, msg)

		// Post-processing
		if err := eventHandler.AfterHandle(ctx.Request.Context(), msg, eventResult); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// []TODO chack if eventResult is impl of our framework error interface, and return response using func from another common package
		if eventResult != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": eventResult.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}
