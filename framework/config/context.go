package config

import (
	"context"
	"net/http"
)

type contextKey struct{}

func FromContext(ctx context.Context) *Config {
	return ctx.Value(contextKey{}).(*Config)
}

func FromRequest(r *http.Request) *Config {
	return FromContext(r.Context())
}

func NewContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

func NewRequest(r *http.Request, cfg *Config) *http.Request {
	return r.WithContext(NewContext(r.Context(), cfg))
}
