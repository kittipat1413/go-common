package event

import (
	"encoding/json"
	"time"
)

const (
	// MetadataKeyVersion is the key for the event message version.
	// It indicates the version of the event schema for compatibility checks.
	MetadataKeyVersion = "version"

	// MetadataKeySource is the key for the event message source.
	// It indicates where the event originated from (e.g., "user-service", "payment-api").
	MetadataKeySource = "source"
)

// EventMessage defines the interface for accessing event data and metadata.
// This interface provides a consistent way to retrieve event information regardless
// of the underlying message format or source (HTTP, message queue, webhook, etc.).
//
// The interface uses generics to provide type-safe access to event payloads while
// maintaining flexibility for different event types and sources.
type EventMessage[T any] interface {
	// GetVersion returns the event schema version for compatibility checks.
	// Used to handle different versions of the same event type.
	GetVersion() string

	// GetEventType returns the type of event (e.g., "user.created", "order.updated").
	// Used for routing and processing logic.
	GetEventType() string

	// GetTimestamp returns when the event occurred.
	// Used for ordering, TTL checks, and temporal processing.
	GetTimestamp() time.Time

	// GetPayload returns the typed event data containing business information.
	// The type T is determined by the EventHandler's generic parameter.
	GetPayload() T

	// GetMetadata returns event metadata and context as key-value pairs.
	// Contains additional information like source, correlation IDs, etc.
	GetMetadata() map[string]string
}

// BaseEventMessage provides a standard implementation of EventMessage[T].
// It supports JSON marshaling/unmarshaling and includes common event fields
// like type, timestamp, payload, and metadata.
//
// This struct can be used directly or embedded in custom event message types
// for consistent event structure across applications.
type BaseEventMessage[T any] struct {
	EventType string            `json:"event_type"` // Type of event (e.g., "user.created")
	Timestamp time.Time         `json:"timestamp"`  // When the event occurred
	Payload   T                 `json:"payload"`    // Typed business data
	Metadata  map[string]string `json:"metadata"`   // Additional context and headers
}

// GetVersion extracts the version from metadata using MetadataKeyVersion.
// Returns empty string if version is not set in metadata.
func (m *BaseEventMessage[T]) GetVersion() string {
	return m.Metadata[MetadataKeyVersion]
}

// GetEventType returns the event type identifier.
// Used for event routing and handler selection.
func (m *BaseEventMessage[T]) GetEventType() string {
	return m.EventType
}

// GetTimestamp returns when the event occurred.
// Essential for event ordering and temporal processing.
func (m *BaseEventMessage[T]) GetTimestamp() time.Time {
	return m.Timestamp
}

// GetPayload returns the typed event business data.
// Contains the actual information relevant to the event.
func (m *BaseEventMessage[T]) GetPayload() T {
	return m.Payload
}

// GetMetadata returns event metadata and context information.
// Includes version, source, correlation IDs, and other contextual data.
func (m *BaseEventMessage[T]) GetMetadata() map[string]string {
	return m.Metadata
}

// BaseEventMessageUnmarshaller unmarshals JSON data into a BaseEventMessage[T].
// This is a convenience function for creating BaseEventMessage instances from JSON bytes.
//
// Parameters:
//   - data: JSON bytes containing event data
//
// Returns:
//   - EventMessage[T]: Unmarshaled event message
//   - error: JSON parsing error if data is invalid
//
// Example:
//
//	data := []byte(`{"event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","payload":{"id":123},"metadata":{"source":"api"}}`)
//	msg, err := BaseEventMessageUnmarshaller[User](data)
func BaseEventMessageUnmarshaller[T any](data []byte) (EventMessage[T], error) {
	var msg BaseEventMessage[T]
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
