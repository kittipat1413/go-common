package event

import (
	"encoding/json"
	"time"
)

const (
	MetadataKeyVersion = "version"
	MetadataKeySource  = "source"
)

type EventMessage[T any] interface {
	GetVersion() string
	GetEventType() string
	GetTimestamp() time.Time
	GetPayload() T
	GetMetadata() map[string]string
}

type BaseEventMessage[T any] struct {
	EventType string            `json:"event_type"`
	Timestamp time.Time         `json:"timestamp"`
	Payload   T                 `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
}

func (m *BaseEventMessage[T]) GetVersion() string {
	return m.Metadata[MetadataKeyVersion]
}

func (m *BaseEventMessage[T]) GetEventType() string {
	return m.EventType
}

func (m *BaseEventMessage[T]) GetTimestamp() time.Time {
	return m.Timestamp
}

func (m *BaseEventMessage[T]) GetPayload() T {
	return m.Payload
}

func (m *BaseEventMessage[T]) GetMetadata() map[string]string {
	return m.Metadata
}

func BaseEventMessageUnmarshaller[T any](data []byte) (EventMessage[T], error) {
	var msg BaseEventMessage[T]
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
