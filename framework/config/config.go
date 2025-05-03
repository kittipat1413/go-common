package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper
}

// Option is a function that configures the Viper instance.
type Option func(v *viper.Viper) error

// WithDefaults injects fallback config values into the Viper instance.
//
// Example:
//
//	config.WithDefaults(map[string]any{
//	  "SERVICE_PORT": ":8080",
//	  "DEBUG_MODE":   true,
//	})
func WithDefaults(defaults map[string]any) Option {
	return func(v *viper.Viper) error {
		for key, val := range defaults {
			v.SetDefault(key, val)
		}
		return nil
	}
}

// WithOptionalConfigPaths loads the first config file found from the given list of paths.
// It will silently skip missing files but return an error if a file is found but unreadable.
//
// Example:
//
//	config.WithOptionalConfigPaths("env.yaml", "../env.yaml")
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

// WithRequiredConfigPath forces the given config file to exist and be readable, or returns an error.
//
// Example:
//
//	config.WithRequiredConfigPath("env.yaml")
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
// It behaves like NewConfig but will terminate the program with a log message if an error occurs.
// This is typically used at application startup when config loading is critical.
//
// Example:
//
//	var AppConfig = config.MustConfig(
//	  config.WithRequiredConfigPath("env.yaml"),
//	  config.WithDefaults(map[string]any{
//	    "SERVICE_PORT": ":8080",
//	  }),
//	)
func MustConfig(opts ...Option) *Config {
	cfg, err := NewConfig(opts...)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return cfg
}

// NewConfig creates and returns a new Config instance using the provided options.
// It returns an error if any option fails.
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

func (c *Config) Get(key string) any { return c.v.Get(key) }

func (c *Config) GetInt(key string) int         { return c.v.GetInt(key) }
func (c *Config) GetBool(key string) bool       { return c.v.GetBool(key) }
func (c *Config) GetString(key string) string   { return c.v.GetString(key) }
func (c *Config) GetFloat64(key string) float64 { return c.v.GetFloat64(key) }

func (c *Config) GetIntSlice(key string) []int       { return c.v.GetIntSlice(key) }
func (c *Config) GetStringSlice(key string) []string { return c.v.GetStringSlice(key) }

func (c *Config) GetStringMap(key string) map[string]any          { return c.v.GetStringMap(key) }
func (c *Config) GetStringMapString(key string) map[string]string { return c.v.GetStringMapString(key) }
func (c *Config) GetStringMapStringSlice(key string) map[string][]string {
	return c.v.GetStringMapStringSlice(key)
}

func (c *Config) GetTime(key string) time.Time         { return c.v.GetTime(key) }
func (c *Config) GetDuration(key string) time.Duration { return c.v.GetDuration(key) }

func (c *Config) All() map[string]interface{} {
	out := map[string]interface{}{}
	for _, key := range c.v.AllKeys() {
		out[key] = c.v.Get(key)
	}
	return out
}
