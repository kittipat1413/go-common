package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/config"
	"github.com/stretchr/testify/assert"
)

func TestWithDefaults(t *testing.T) {
	cfg := config.MustConfig(
		config.WithDefaults(map[string]any{
			"FOO":     "bar",
			"INT_VAL": 42,
		}),
	)

	assert.Equal(t, "bar", cfg.GetString("FOO"))
	assert.Equal(t, 42, cfg.GetInt("INT_VAL"))
}

func TestWithOptionalConfigPaths_FileExists(t *testing.T) {
	// Arrange: Create a temporary config file
	tmpFile := createTempYamlFile(t, `
MY_KEY: hello
MY_NUMBER: 123
`)
	defer os.Remove(tmpFile)

	cfg := config.MustConfig(
		config.WithOptionalConfigPaths(tmpFile),
	)

	assert.Equal(t, "hello", cfg.GetString("MY_KEY"))
	assert.Equal(t, 123, cfg.GetInt("MY_NUMBER"))
}

func TestWithOptionalConfigPaths_FileMissing(t *testing.T) {
	cfg := config.MustConfig(
		config.WithOptionalConfigPaths("nonexistent.yaml"),
		config.WithDefaults(map[string]any{
			"DEFAULT_KEY": "fallback",
		}),
	)

	assert.Equal(t, "fallback", cfg.GetString("DEFAULT_KEY"))
}

func TestWithRequiredConfigPath_FileExists(t *testing.T) {
	tmpFile := createTempYamlFile(t, `
REQUIRED_KEY: value
`)
	defer os.Remove(tmpFile)

	cfg := config.MustConfig(
		config.WithRequiredConfigPath(tmpFile),
	)

	assert.Equal(t, "value", cfg.GetString("REQUIRED_KEY"))
}

func TestWithRequiredConfigPath_FileMissing(t *testing.T) {
	_, err := config.NewConfig(
		config.WithRequiredConfigPath("does_not_exist.yaml"),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required config file not found")
}

func TestAllGetters(t *testing.T) {
	now := time.Now().Truncate(time.Second) // Remove nanoseconds for comparison

	defaults := map[string]any{
		"ANY":              "value",
		"INT_VAL":          123,
		"BOOL_VAL":         true,
		"STRING_VAL":       "hello",
		"FLOAT_VAL":        3.1415,
		"INT_SLICE":        []int{1, 2, 3},
		"STRING_SLICE":     []string{"a", "b", "c"},
		"MAP_ANY":          map[string]any{"k": "v"},
		"MAP_STRING":       map[string]string{"x": "y"},
		"MAP_STRING_SLICE": map[string][]string{"group": {"one", "two"}},
		"TIME_VAL":         now,
		"DURATION_VAL":     "1h30m",
	}

	cfg := config.MustConfig(
		config.WithDefaults(defaults),
	)

	assert.Equal(t, "value", cfg.Get("ANY"))
	assert.Equal(t, 123, cfg.GetInt("INT_VAL"))
	assert.Equal(t, true, cfg.GetBool("BOOL_VAL"))
	assert.Equal(t, "hello", cfg.GetString("STRING_VAL"))
	assert.InDelta(t, 3.1415, cfg.GetFloat64("FLOAT_VAL"), 0.0001)

	assert.Equal(t, []int{1, 2, 3}, cfg.GetIntSlice("INT_SLICE"))
	assert.Equal(t, []string{"a", "b", "c"}, cfg.GetStringSlice("STRING_SLICE"))
	assert.Equal(t, map[string]any{"k": "v"}, cfg.GetStringMap("MAP_ANY"))
	assert.Equal(t, map[string]string{"x": "y"}, cfg.GetStringMapString("MAP_STRING"))
	assert.Equal(t, map[string][]string{"group": {"one", "two"}}, cfg.GetStringMapStringSlice("MAP_STRING_SLICE"))

	assert.Equal(t, now, cfg.GetTime("TIME_VAL"))

	expectedDuration, _ := time.ParseDuration("1h30m")
	assert.Equal(t, expectedDuration, cfg.GetDuration("DURATION_VAL"))
}

func TestAllKeys(t *testing.T) {
	cfg := config.MustConfig(
		config.WithDefaults(map[string]any{
			"a": 1,
			"b": 2,
			"c": "three",
		}),
	)

	all := cfg.All()
	assert.Equal(t, 1, all["a"])
	assert.Equal(t, 2, all["b"])
	assert.Equal(t, "three", all["c"])
}

func createTempYamlFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp yaml file: %v", err)
	}

	return path
}
