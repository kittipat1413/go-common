# Event Package
This directory contains the **Event Handling Framework** for Go applications, designed to streamline event-driven workflows in backend services. It provides a generic and modular way to handle event messages with flexible payloads and versioning support.

## Introduction
The Event Package provides a robust framework for handling event messages in Go applications. It allows for:
- Defining and processing event messages with flexible, user-defined payload types.
- Integration with HTTP frameworks (like Gin) via a generic EventHandler interface.
- Separation of concerns: event message parsing, business logic handling, and HTTP integration are modular and distinct.
- This package promotes reusable patterns across services to maintain consistency in event handling across your projects.

## Features
- **User-Defined Payload Types:** Utilizes Go generics to allow flexible payload structures for event messages.
- **Generic Interface for Event Processing:** Provides a flexible `EventHandler` interface for defining how events are processed.
- **Modular Design:** Separates event message logic, event handler logic, and HTTP integration for clean and maintainable code.
- **HTTP Integration:** Includes helper functions for integrating with popular web frameworks like Gin.

## Usage

### Defining Event Messages
To define an event, the framework expects a JSON structure with an event type, timestamp, and a flexible payload. The `payload` section of the event can be a user-defined structure, providing the flexibility to handle different kinds of events.

example
```json
{
  "event_type": "user_created",
  "timestamp": "2024-08-20T05:21:19.143357839Z",
  "payload": {
    "user_id": "12345",
    "username": "johndoe"
  },
  "metadata": {
    "source": "user_service",
    "version": "1.0"
  }
}
```
- `event_type`: A string indicating the type of event.
- `timestamp`: An RFC3339 formatted timestamp.
- `payload`: An object containing the event data. This is where generics come into play, allowing for flexible payload types.
- `metadata`: An object containing metadata about the event, including the `version`.

### Defining Your Payload
Using Go's generics, you can define your own payload type to represent the data for each event:
```golang
// event_payload.go
package your_package

type UserCreatedPayload struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
}
```

### Creating an Event Handler
The event framework uses a generic `EventHandler` interface to process events.
```golang
type EventHandler[T any] interface {
	BeforeHandle(ctx context.Context, msg EventMessage[T]) error
	AfterHandle(ctx context.Context, msg EventMessage[T], eventResult error) error
	UnmarshalEventMessage(data []byte) (EventMessage[T], error)
}
```
> The event framework is designed to be extensible. You can create custom event handlers to handle specific scenarios, such as events with callbacks, or complex retry mechanisms. For examples of custom handlers, see: [custom_handler/callback/handler.go](custom_handler/callback/handler.go)

### Implementing Business Logic
Define your business logic function that processes the event payload.
```golang
// handlers.go
package your_package

import (
    "github.com/gin-gonic/gin"
    "https://github.com/kittipat1413/go-common/framework/event"
)

func MyBusinessLogic(ctx *gin.Context, msg event.EventMessage[UserCreatedPayload]) error {
    payload := msg.GetPayload()

    // Process the event
    log.Printf("User created: %s with ID %s", payload.Username, payload.UserID)

    // Return nil on success, or an error if something goes wrong
    return nil
}
```

### Integrating with HTTP Frameworks (e.g., Gin)
If you are using a web framework like Gin, you can integrate the event framework with the `NewGinEventHandler`, which creates a `gin.HandlerFunc`. This handler can then be injected into your Gin routes. The handler requires an `EventHandler` instance to process the event.
```golang
// main.go

import (
    "github.com/gin-gonic/gin"
    "https://github.com/kittipat1413/go-common/framework/handler"
)

func main() {
    router := gin.Default()

    // Initialize EventHandler as before

    // Create the Gin handler
    ginHandler := handler.NewGinEventHandler(MyBusinessLogic, eventHandler)

    // Register the route with the Gin router
    router.POST("/event", ginHandler)

    // Optionally, register a route to handle callbacks
    router.GET("/callback", CallbackHandler)

    router.Run(":8080")
}
```

## Example
You can find a complete working example in the repository under [framework/event/example](example/).
