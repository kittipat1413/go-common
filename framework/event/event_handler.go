package event

import "context"

//go:generate mockgen -source=./event_handler.go -destination=./mocks/event_handler.go -package=event_handler_mocks

// EventHandler defines the interface for processing events with lifecycle hooks.
// Implementations provide custom logic for event unmarshaling, pre-processing,
// and post-processing while maintaining consistent event handling patterns.
//
// Lifecycle Flow:
//  1. UnmarshalEventMessage() - Convert raw data to typed event message
//  2. BeforeHandle() - Pre-processing and validation
//  3. [External event processing logic]
//  4. AfterHandle() - Post-processing and cleanup
type EventHandler[T any] interface {
	// BeforeHandle performs pre-processing before the main event handling logic.
	// Used for validation, authentication, enrichment, or preparation tasks.
	BeforeHandle(ctx context.Context, msg EventMessage[T]) error

	// AfterHandle performs post-processing after the main event handling logic.
	// Runs regardless of main processing success/failure, suitable for cleanup,
	// callbacks, logging, and side effects.
	AfterHandle(ctx context.Context, msg EventMessage[T], eventResult error) error

	// UnmarshalEventMessage converts raw event data into a typed EventMessage.
	// Handles deserialization from various formats (JSON, XML, protobuf, etc.)
	// and returns strongly-typed event message for processing.
	UnmarshalEventMessage(data []byte) (EventMessage[T], error)
}
