package config_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kittipat1413/go-common/framework/config"
	"github.com/stretchr/testify/assert"
)

func TestContextInjection(t *testing.T) {
	// Arrange
	cfg := config.MustConfig(config.WithDefaults(map[string]any{
		"FOO": "bar",
	}))

	// Act
	ctx := context.Background()
	ctxWithCfg := config.NewContext(ctx, cfg)
	retrievedCfg := config.FromContext(ctxWithCfg)

	// Assert
	assert.NotNil(t, retrievedCfg)
	assert.Equal(t, "bar", retrievedCfg.GetString("FOO"))
}

func TestRequestInjection(t *testing.T) {
	// Arrange
	cfg := config.MustConfig(config.WithDefaults(map[string]any{
		"FOO": "baz",
	}))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	reqWithCfg := config.NewRequest(req, cfg)

	// Act
	retrievedCfg := config.FromRequest(reqWithCfg)

	// Assert
	assert.NotNil(t, retrievedCfg)
	assert.Equal(t, "baz", retrievedCfg.GetString("FOO"))
}
