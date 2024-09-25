package handler

import (
	"net/http"

	"github.com/kittipat1413/go-common/framework/event"

	"github.com/gin-gonic/gin"
)

// GinEventHandlerFunc defines a generic function type for handling events in Gin
type GinEventHandlerFunc[T any] func(ctx *gin.Context, msg event.EventMessage[T]) error

// NewGinEventHandler is a method of EventHandler that creates a Gin-specific handler
func NewGinEventHandler[T any](ginHandler GinEventHandlerFunc[T], eventHandler *event.EventHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Read the raw JSON data
		data, err := ctx.GetRawData()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		msg, err := event.UnmarshalEventMessage[T](data)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Execute the business logic with the typed payload
		handlerErr := ginHandler(ctx, msg)

		// Dispatch based on the version
		switch msg := msg.(type) {
		case *event.EventMessageV1[T]:
			eventHandler.HandleCallback(handlerErr, msg.Callback)
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event message version"})
			return
		}

		if handlerErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": handlerErr.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}
