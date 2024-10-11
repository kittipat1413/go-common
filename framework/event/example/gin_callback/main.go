package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kittipat1413/go-common/framework/event"
	callbackhandler "github.com/kittipat1413/go-common/framework/event/custom_handler/callback"
	httphandler "github.com/kittipat1413/go-common/framework/event/http_handler"

	"github.com/gin-gonic/gin"
)

/*
To run this example, use the following curl command:

- with callback
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
          "event_type": "test_event",
          "timestamp": "2024-08-20T05:21:19.143357839Z",
          "payload": {
            "message": "Hello, World!",
            "value": 42
          },
          "metadata": {
            "source": "test_source",
            "version": "1.0"
          },
          "callback": {
            "success_url": "http://localhost:8080/callback",
            "fail_url": "http://localhost:8080/callback"
          }
        }' \
     http://localhost:8080/event

- without callback
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
          "event_type": "test_event",
          "timestamp": "2024-08-20T05:21:19.143357839Z",
          "payload": {
            "message": "Hello, World!",
            "value": 42
          },
          "metadata": {
            "source": "test_source",
            "version": "1.0"
          }
        }' \
     http://localhost:8080/event
*/

func main() {
	// Initialize the Gin router
	router := gin.Default()

	// Register the route with the Gin router
	router.POST("/event", eventHandler())

	// Register the route for handling callbacks
	router.GET("/callback", callbackHandler)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}

// Define your payload type
type MyPayload struct {
	Message string `json:"message"`
	Value   int    `json:"value"`
}

// Handler for User Creation Event
func eventHandler() gin.HandlerFunc {
	// Create a custom http.Client if needed
	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}
	eventHandler := callbackhandler.NewEventHandler(
		callbackhandler.WithHTTPClient[MyPayload](httpClient),
		callbackhandler.WithCallbackConfig[MyPayload](2, 1*time.Second, 1*time.Minute),
	)
	return httphandler.NewGinEventHandler(businessLogic, eventHandler)
}
func businessLogic(ctx *gin.Context, msg event.EventMessage[MyPayload]) error {
	// Access the payload
	payload := msg.GetPayload()

	// Implement your business logic here
	// For this example, we'll simply log the payload
	fmt.Printf("Received payload: %+v\n", payload)

	// Optionally, perform operations based on the payload
	// Return nil if successful, or an error if something went wrong
	return nil
}

// callbackHandler handles incoming callback requests
func callbackHandler(ctx *gin.Context) {
	fmt.Printf("Received callback request -> Method: %s, URL: %s\n", ctx.Request.Method, ctx.Request.RequestURI)

	// You can access query parameters or headers if needed
	// For example, to get a query parameter named "status":
	// status := ctx.Query("status")
	// fmt.Printf("Status: %s\n", status)

	// Send a response back to the callback sender
	ctx.JSON(http.StatusOK, gin.H{"message": "Callback received successfully"})
}
