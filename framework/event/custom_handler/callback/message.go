package callbackhandler

import "github.com/kittipat1413/go-common/framework/event"

type CallbackEventMessage[T any] struct {
	event.BaseEventMessage[T]
	Callback *CallbackInfo `json:"callback"`
}

type CallbackInfo struct {
	SuccessURL string `json:"success_url"`
	FailURL    string `json:"fail_url"`
}

func (m *CallbackEventMessage[T]) GetCallback() *CallbackInfo {
	return m.Callback
}
