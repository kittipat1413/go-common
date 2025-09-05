package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config wraps a Viper instance to provide a simplified interface for configuration management.
// It offers type-safe methods for retrieving configuration values and handles the underlying
// Viper complexity.
type Config struct {
	v *viper.Viper
}

// Option is a function that configures the Viper instance during initialization.
// Options are applied in the order they are provided to NewConfig or MustConfig.
type Option func(v *viper.Viper) error

// WithDefaults injects fallback configuration values into the Viper instance.
// These values are used when no value is found in config files or environment variables.
//
// Default values should use the same key names as expected in config files and environment variables.
//
// Example:
//
//	config := MustConfig(
//		WithDefaults(map[string]any{
//			"SERVICE_PORT": ":8080",
//			"DEBUG_MODE":   false,
//			"DATABASE.TIMEOUT": "30s",
//			"REDIS.POOL_SIZE": 10,
//		}),
//	)
func WithDefaults(defaults map[string]any) Option {
	return func(v *viper.Viper) error {
		for key, val := range defaults {
			v.SetDefault(key, val)
		}
		return nil
	}
}

// WithOptionalConfigPaths attempts to load the first configuration file found from the given list of paths.
// It will silently skip missing files but return an error if a file exists but cannot be read or parsed.
//
// Example:
//
//	// Try local config first, then fallback to default location
//	config := MustConfig(
//		WithOptionalConfigPaths(
//			"./local.env.yaml",
//			"./config/env.yaml",
//			"/etc/myapp/config.yaml",
//		),
//	)
func WithOptionalConfigPaths(paths ...string) Option {
	return func(v *viper.Viper) error {
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				v.SetConfigFile(path)
				if err := v.ReadInConfig(); err != nil {
					return fmt.Errorf("failed to read optional config file %s: %w", path, err)
				}
				break
			}
		}
		return nil
	}
}

// WithRequiredConfigPath forces the specified configuration file to exist and be readable.
// If the file doesn't exist or cannot be read, an error is returned.
//
// Example:
//
//	config := MustConfig(
//		WithRequiredConfigPath("./config/production.yaml"),
//	)
func WithRequiredConfigPath(path string) Option {
	return func(v *viper.Viper) error {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("required config file not found: %s", path)
		}
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file %s: %w", path, err)
		}
		return nil
	}
}

// MustConfig creates and returns a new Config instance using the provided options.
// It behaves like NewConfig but will terminate the program with log.Fatalln if any error occurs.
//
// This function is typically used at application startup when configuration loading is critical
// and the application cannot continue without proper configuration.
//
// Warning: This function calls log.Fatalln on error, which will terminate the program.
// Use NewConfig if you need error handling instead of program termination.
//
// Example:
//
//	// Global config that must load or app fails to start
//	var AppConfig = config.MustConfig(
//		config.WithRequiredConfigPath("env.yaml"),
//		config.WithDefaults(map[string]any{
//			"SERVICE_PORT": ":8080",
//			"DEBUG_MODE":   false,
//		}),
//	)
//
//	func main() {
//		port := AppConfig.GetString("SERVICE_PORT")
//		// ... rest of application
//	}
func MustConfig(opts ...Option) *Config {
	cfg, err := NewConfig(opts...)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return cfg
}

// NewConfig creates and returns a new Config instance using the provided options.
// It initializes Viper with automatic environment variable reading and applies all options in order.
//
// Environment variable reading is automatically enabled, meaning any environment variable
// can be accessed using its name as a key.
//
// Example:
//
//	config, err := NewConfig(
//		WithOptionalConfigPaths("config.yaml"),
//		WithDefaults(map[string]any{
//			"SERVICE_PORT": ":8080",
//		}),
//	)
//	if err != nil {
//		log.Fatalf("Failed to load config: %v", err)
//	}
func NewConfig(opts ...Option) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, err
		}
	}

	return &Config{v: v}, nil
}

// Get retrieves the value associated with the key as an interface{}.
func (c *Config) Get(key string) any { return c.v.Get(key) }

// GetInt retrieves the value associated with the key as an integer.
func (c *Config) GetInt(key string) int { return c.v.GetInt(key) }

// GetBool retrieves the value associated with the key as a boolean.
func (c *Config) GetBool(key string) bool { return c.v.GetBool(key) }

// GetString retrieves the value associated with the key as a string.
func (c *Config) GetString(key string) string { return c.v.GetString(key) }

// GetFloat64 retrieves the value associated with the key as a float64.
func (c *Config) GetFloat64(key string) float64 { return c.v.GetFloat64(key) }

// GetIntSlice retrieves the value associated with the key as a slice of integers.
func (c *Config) GetIntSlice(key string) []int { return c.v.GetIntSlice(key) }

// GetStringSlice retrieves the value associated with the key as a slice of strings.
func (c *Config) GetStringSlice(key string) []string { return c.v.GetStringSlice(key) }

// GetStringMap retrieves the value associated with the key as a map[string]any.
func (c *Config) GetStringMap(key string) map[string]any { return c.v.GetStringMap(key) }

// GetStringMapString retrieves the value associated with the key as a map[string]string.
func (c *Config) GetStringMapString(key string) map[string]string { return c.v.GetStringMapString(key) }

// GetStringMapStringSlice retrieves the value associated with the key as a map[string][]string.
func (c *Config) GetStringMapStringSlice(key string) map[string][]string {
	return c.v.GetStringMapStringSlice(key)
}

// GetTime retrieves the value associated with the key as a time.Time.
func (c *Config) GetTime(key string) time.Time { return c.v.GetTime(key) }

// GetDuration retrieves the value associated with the key as a time.Duration.
func (c *Config) GetDuration(key string) time.Duration { return c.v.GetDuration(key) }

// All returns a map containing all configuration key-value pairs.
// This includes values from all sources (defaults, config files, environment variables).
//
// Keys are normalized to lowercase. This is useful for debugging configuration
// or when you need to iterate over all available configuration values.
func (c *Config) All() map[string]interface{} {
	out := map[string]interface{}{}
	for _, key := range c.v.AllKeys() {
		out[key] = c.v.Get(key)
	}
	return out
}
