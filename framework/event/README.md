# Event Package
This directory contains the Event Handling Framework for Go applications, designed to streamline the processing of event-driven workflows in backend services. It is a part of a larger collection of utilities aimed at standardizing backend API repositories within our organization.

## Introduction
The Event Package provides a robust and flexible framework for handling event messages in Go applications. It supports:
- Parsing and validating event messages with version control.
- Applying consistent business logic processing across services.
- Handling success and failure callbacks with configurable retry mechanisms.
- This package is part of our central codebase aimed at unifying code standards across all backend API projects, promoting reusability and maintainability.

## Features
- Versioned Event Parsing: Supports multiple versions of event message formats for forward compatibility.
- Generic Payload Handling: Utilizes Go generics for flexible payload structures.
- Callback Handling with Retries: Manages success and failure callbacks with exponential backoff and jitter to handle transient failures.
- Configurable Options: Leverages the functional options pattern for customizable configurations.
- Integration with Gin: Provides middleware and handler functions compatible with the Gin framework.

## Usage

### Defining Event Messages
Event messages are expected to be in JSON format with a specific structure. Here's a simple

example
```json
{
  "event_type": "{event_type_name}",
  "timestamp": "2024-08-20T05:21:19.143357839Z",
  "payload": { /* any */ },
  "metadata": {
    "source": "{source_name}",
    "version": "1.0"
  },
  "callback": {
    "success_url": "{success_url}",
    "fail_url": "{fail_url}"
  }
}
```
- `event_type`: A string indicating the type of event.
- `timestamp`: An RFC3339 formatted timestamp.
- `payload`: An object containing the event data. This is where generics come into play, allowing for flexible payload types.
- `metadata`: An object containing metadata about the event, including the `version`.
- `callback`: An object specifying URLs for success and failure callbacks (optional).

The use of Go's generics allows the framework to handle events with different payload types while maintaining type safety and reducing code duplication.
```golang
// event_payload.go
package your_package

type MyEventPayload struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}
```

### Creating an Event Handler
Create an instance of EventHandler, optionally configuring it with custom settings using the functional options pattern.
```golang
// main.go
package main

import (
    "net/http"
    "time"

    "https://github.com/kittipat1413/go-common/framework/event"
)

func main() {
    // Initialize EventHandler with custom configurations
    eventHandler := event.NewEventHandler(
        event.WithHTTPClient(&http.Client{
            Timeout: 15 * time.Second,
        }),
        event.WithCallbackConfig(5, 2*time.Second, 1*time.Minute),
    )

    // ... rest of your application setup
}
```

### Implementing Business Logic
Define your business logic function that processes the event payload.
```golang
// handlers.go
package your_package

import (
    "github.com/gin-gonic/gin"
    "https://github.com/kittipat1413/go-common/framework/event"
)

func MyBusinessLogic(ctx *gin.Context, msg event.EventMessage[MyEventPayload]) error {
    payload := msg.GetPayload()
    // Implement your processing logic here
    // Return nil on success or an error on failure
    return nil
}
```
### Setting Up Gin Routes
Integrate the event handler into your Gin routes.
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


## Configuration
The `EventHandler` can be configured using the functional options pattern. Available options include:
- `WithHTTPClient`: Sets a custom http.Client for making callback requests.
- `WithCallbackConfig`: Sets the callback retry parameters.
Default Configuration

If no options are provided, the following defaults are used:
- `HTTP Client`: http.DefaultClient
- `Max Retries`: 3
- `Retry Interval`: 2 * time.Second
- `Callback Timeout`: 1 * time.Minute