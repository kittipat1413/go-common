package config

import (
	"context"
	"net/http"
)

// contextKey is a private struct type used as a key for storing Config in context.Context.
type contextKey struct{}

// FromContext retrieves a Config instance from the provided context.Context.
//
// Example:
//
//	cfg := config.FromContext(ctx)
func FromContext(ctx context.Context) *Config {
	return ctx.Value(contextKey{}).(*Config)
}

// FromRequest retrieves a Config instance from an http.Request's context.
//
// Example:
//
//	cfg := config.FromRequest(r)
func FromRequest(r *http.Request) *Config {
	return FromContext(r.Context())
}

// NewContext creates a new context.Context with the provided Config instance attached.
//
// Example:
//
//	// Inject into root context
//	ctx := config.NewContext(context.Background(), appConfig)
//	// Use ctx in your application ...
func NewContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

// NewRequest creates a new http.Request with the Config instance stored in its context.
//
// Example:
//
//	// Inject config into request
//	r = config.NewRequest(r, cfg)
//	// Continue with enhanced request ...
func NewRequest(r *http.Request, cfg *Config) *http.Request {
	return r.WithContext(NewContext(r.Context(), cfg))
}
