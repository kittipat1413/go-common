package httphandler

import (
	"net/http"

	"github.com/kittipat1413/go-common/framework/event"

	"github.com/gin-gonic/gin"
)

// GinEventHandlerFunc is a function that handles a business logic of an event message
type GinEventHandlerFunc[T any] func(ctx *gin.Context, msg event.EventMessage[T]) error

// NewGinEventHandler function creates a new Gin HTTP handler function that processes incoming JSON data, unmarshals it into an event message, executes business logic, and handles the event using a provided event handler.
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
		if err := eventHandler.BeforeHandle(ctx, msg); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Execute the business logic with the event message
		eventResult := ginHandler(ctx, msg)

		// Post-processing
		if err := eventHandler.AfterHandle(ctx, msg, eventResult); err != nil {
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
