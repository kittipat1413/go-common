package event_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kittipat1413/go-common/framework/event"
)

func TestBaseEventMessageMethods(t *testing.T) {
	type SamplePayload struct {
		Data string `json:"data"`
	}
	timestamp := time.Now()
	metadata := map[string]string{
		event.MetadataKeyVersion: "1.0",
		event.MetadataKeySource:  "unit_test",
	}
	payload := SamplePayload{Data: "test data"}

	msg := &event.BaseEventMessage[SamplePayload]{
		EventType: "test_event",
		Timestamp: timestamp,
		Payload:   payload,
		Metadata:  metadata,
	}

	assert.Equal(t, "1.0", msg.GetVersion())
	assert.Equal(t, "test_event", msg.GetEventType())
	assert.Equal(t, timestamp, msg.GetTimestamp())
	assert.Equal(t, payload, msg.GetPayload())
	assert.Equal(t, metadata, msg.GetMetadata())
}

func TestBaseEventMessageUnmarshaller(t *testing.T) {
	type SamplePayload struct {
		Data string `json:"data"`
	}
	type testCase struct {
		name        string
		inputJSON   string
		expectError bool
		expectedMsg *event.BaseEventMessage[SamplePayload]
	}

	timestamp := time.Now().UTC()

	testCases := []testCase{
		{
			name: "Valid Message",
			inputJSON: fmt.Sprintf(`{
				"event_type": "test_event",
				"timestamp": "%s",
				"payload": {"data": "test data"},
				"metadata": {"version": "1.0", "source": "unit_test"}
			}`, timestamp.Format(time.RFC3339Nano)),
			expectError: false,
			expectedMsg: &event.BaseEventMessage[SamplePayload]{
				EventType: "test_event",
				Timestamp: timestamp,
				Payload:   SamplePayload{Data: "test data"},
				Metadata: map[string]string{
					event.MetadataKeyVersion: "1.0",
					event.MetadataKeySource:  "unit_test",
				},
			},
		},
		{
			name:        "Invalid JSON",
			inputJSON:   `{"event_type": "test_event", "timestamp": "invalid_timestamp"`,
			expectError: true,
		},
		{
			name: "Invalid Timestamp Format",
			inputJSON: `{
				"event_type": "test_event",
				"timestamp": "not_a_timestamp",
				"payload": {"data": "test data"},
				"metadata": {"version": "1.0", "source": "unit_test"}
			}`,
			expectError: true,
		},
		{
			name: "Invalid Payload Type",
			inputJSON: fmt.Sprintf(`{
				"event_type": "test_event",
				"timestamp": "%s",
				"payload": {"data": 123},
				"metadata": {"version": "1.0", "source": "unit_test"}
			}`, timestamp.Format(time.RFC3339Nano)),
			expectError: true,
		},
		{
			name: "Valid Message with Extra Fields",
			inputJSON: fmt.Sprintf(`{
				"event_type": "test_event",
				"timestamp": "%s",
				"payload": {"data": "test data", "extra": "value"},
				"metadata": {"version": "1.0", "source": "unit_test", "extra_meta": "meta_value"},
				"extra_field": "extra_value"
			}`, timestamp.Format(time.RFC3339Nano)),
			expectError: false,
			expectedMsg: &event.BaseEventMessage[SamplePayload]{
				EventType: "test_event",
				Timestamp: timestamp,
				Payload:   SamplePayload{Data: "test data"}, // Extra fields are ignored
				Metadata: map[string]string{
					event.MetadataKeyVersion: "1.0",
					event.MetadataKeySource:  "unit_test",
					"extra_meta":             "meta_value",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Unmarshal the JSON data using the BaseEventMessageUnmarshaller
			unmarshalledMsg, err := event.BaseEventMessageUnmarshaller[SamplePayload]([]byte(tc.inputJSON))

			if tc.expectError {
				assert.Error(t, err, "Expected an error for test case: %s", tc.name)
				return
			}

			assert.NoError(t, err, "Did not expect an error for test case: %s", tc.name)
			assert.NotNil(t, unmarshalledMsg)

			// Assert that the unmarshalled message has the same content
			assert.Equal(t, tc.expectedMsg.GetVersion(), unmarshalledMsg.GetVersion())
			assert.Equal(t, tc.expectedMsg.GetEventType(), unmarshalledMsg.GetEventType())
			assert.Equal(t, tc.expectedMsg.GetTimestamp(), unmarshalledMsg.GetTimestamp())
			assert.Equal(t, tc.expectedMsg.GetPayload(), unmarshalledMsg.GetPayload())
			assert.Equal(t, tc.expectedMsg.GetMetadata(), unmarshalledMsg.GetMetadata())
		})
	}
}
