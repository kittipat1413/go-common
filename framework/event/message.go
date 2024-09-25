package event

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	MetadataKeyVersion = "version"
	MetadataKeySource  = "source"
)

const (
	EventV1 = "1.0"
)

type EventMessage[T any] interface {
	GetVersion() string
	GetEventType() string
	GetPayload() T
}

type BaseEventMessage[T any] struct {
	EventType string            `json:"event_type"`
	Timestamp time.Time         `json:"timestamp"`
	Payload   T                 `json:"payload"`
	Metadata  map[string]string `json:"metadata"`
}

func (b *BaseEventMessage[T]) GetVersion() string {
	return b.Metadata[MetadataKeyVersion]
}

func (b *BaseEventMessage[T]) GetEventType() string {
	return b.EventType
}

func (b *BaseEventMessage[T]) GetPayload() T {
	return b.Payload
}

type EventMessageV1[T any] struct {
	BaseEventMessage[T]
	Callback *CallbackInfo `json:"callback"`
}

type CallbackInfo struct {
	SuccessURL string `json:"success_url"`
	FailURL    string `json:"fail_url"`
}

func UnmarshalEventMessage[T any](data []byte) (EventMessage[T], error) {
	// Unmarshal into a map to extract metadata and version
	var rawMsg map[string]interface{}
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return nil, err
	}

	// Extract metadata
	metadataValue, ok := rawMsg["metadata"]
	if !ok {
		return nil, fmt.Errorf("missing metadata field")
	}
	metadata, ok := metadataValue.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("metadata field is not a map")
	}

	// Extract version
	versionValue, ok := metadata[MetadataKeyVersion]
	if !ok {
		return nil, fmt.Errorf("missing version in metadata")
	}
	version, ok := versionValue.(string)
	if !ok {
		return nil, fmt.Errorf("version is not a string")
	}

	// Switch based on version
	switch version {
	case EventV1:
		// Unmarshal into EventMessageV1[T]
		var msg EventMessageV1[T]
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	default:
		return nil, fmt.Errorf("unsupported version: %s", version)
	}
}
