package event

import "context"

//go:generate mockgen -source=./event_handler.go -destination=./mocks/event_handler.go -package=event_handler_mocks
type EventHandler[T any] interface {
	BeforeHandle(ctx context.Context, msg EventMessage[T]) error
	AfterHandle(ctx context.Context, msg EventMessage[T], eventResult error) error
	UnmarshalEventMessage(data []byte) (EventMessage[T], error)
}
