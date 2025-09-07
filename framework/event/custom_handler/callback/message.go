package callbackhandler

import "github.com/kittipat1413/go-common/framework/event"

// CallbackEventMessage extends BaseEventMessage with HTTP callback functionality.
// It embeds the standard event message structure and adds callback URL configuration
// for sending notifications based on event processing results.
//
// JSON Structure:
//
//	{
//	    "event_type": "user.created",
//	    "timestamp": "2023-01-01T12:00:00Z",
//	    "payload": { /* your event data */ },
//	    "metadata": { "source": "api" },
//	    "callback": {
//	        "success_url": "https://api.example.com/success",
//	        "fail_url": "https://api.example.com/failure"
//	    }
//	}
type CallbackEventMessage[T any] struct {
	event.BaseEventMessage[T]
	Callback *CallbackInfo `json:"callback"`
}

// CallbackInfo contains HTTP callback URLs for event processing notifications.
// Supports separate URLs for success and failure scenarios, enabling flexible
// webhook delivery patterns based on processing outcomes.
type CallbackInfo struct {
	SuccessURL string `json:"success_url"` // URL called when event processing succeeds
	FailURL    string `json:"fail_url"`    // URL called when event processing fails
}

// GetCallback returns the callback configuration for this event message.
// Used by the callback handler to determine which URLs to call based on
// event processing results.
func (m *CallbackEventMessage[T]) GetCallback() *CallbackInfo {
	return m.Callback
}
