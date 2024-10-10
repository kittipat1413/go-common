package logger

import (
	"context"
	"net/http"
)

type contextKey struct{}

var loggerKey = &contextKey{}

// FromContext retrieves the Logger from the context. It returns a default logger if the context doesn't have one.
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return NewDefaultLogger()
}

// FromRequest retrieves the Logger from the HTTP request's context.
func FromRequest(r *http.Request) Logger {
	return FromContext(r.Context())
}

// NewContext returns a new Context that carries the provided Logger.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// NewRequest returns a new *http.Request that carries the provided Logger.
func NewRequest(r *http.Request, logger Logger) *http.Request {
	return r.WithContext(NewContext(r.Context(), logger))
}
