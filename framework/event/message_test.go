package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalEventMessage(t *testing.T) {
	// Define the payload types for testing
	type TestPayloadV1 struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	// Define test cases in a table-driven approach
	tests := []struct {
		name            string
		jsonData        string
		expectedPayload interface{}
		expectedError   string
	}{
		{
			name: "Valid V1 Message",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {
					"field1": "value1",
					"field2": 42
				},
				"metadata": {
					"source": "test_source",
					"version": "1.0"
				},
				"callback": {
					"success_url": "http://success.url",
					"fail_url": "http://fail.url"
				}
			}`,
			expectedPayload: TestPayloadV1{
				Field1: "value1",
				Field2: 42,
			},
		},
		{
			name: "Type Mismatch in Payload",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {
					"field2": "should_be_int"
				},
				"metadata": {
					"source": "test_source",
					"version": "1.0"
				}
			}`,
			expectedError: "cannot unmarshal string into Go struct field",
		},
		{
			name: "Extra Fields in Payload",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {
					"field1": "value1",
					"extra_field": "extra_value"
				},
				"metadata": {
					"source": "test_source",
					"version": "1.0"
				}
			}`,
			expectedPayload: TestPayloadV1{
				Field1: "value1",
				Field2: 0, // Field2 is missing, so it should be zero value
			},
		},
		{
			name: "Unsupported Version",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {},
				"metadata": {
					"source": "test_source",
					"version": "2.0"
				}
			}`,
			expectedError: "unsupported version: 2.0",
		},
		{
			name: "Missing Version",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {},
				"metadata": {
					"source": "test_source"
				}
			}`,
			expectedError: "missing version in metadata",
		},
		{
			name: "Invalid JSON",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {},
				"metadata": {
					"source": "test_source",
					"version": "1.0"
				},
				"callback": {
					"success_url": "http://success.url",
					"fail_url": "http://fail.url"
				}
			`, // Missing closing brace
			expectedError: "unexpected end of JSON input",
		},
		{
			name: "Missing Metadata",
			jsonData: `{
				"event_type": "test_event",
				"timestamp": "2024-08-20T05:21:19.143357839Z",
				"payload": {}
			}`,
			expectedError: "missing metadata field",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Determine the payload type based on the test case
			switch tc.name {
			case "Valid V1 Message", "Type Mismatch in Payload", "Extra Fields in Payload":
				// For V1 messages, use TestPayloadV1
				msg, err := UnmarshalEventMessage[TestPayloadV1]([]byte(tc.jsonData))

				if tc.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, msg)

				// Assert the payload
				assert.Equal(t, tc.expectedPayload, msg.GetPayload())

			default:
				// For other test cases, use interface{} for payload
				msg, err := UnmarshalEventMessage[interface{}]([]byte(tc.jsonData))

				if tc.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, msg)
			}
		})
	}
}
