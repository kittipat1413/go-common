[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Config Package
The config package provides a reusable and extensible configuration loader built on top of `Viper`. It is designed to simplify configuration management in Go applications by supporting multiple config sources, context propagation, and functional loading options.

## Features
- **Environment-First Configuration**: Automatically reads from environment variables.
- **YAML File Support**: Load config from `.yaml` files (required or optional).
- **Injectable Defaults**: Provide fallback values when env or file values are not present.
- **Context Integration**: Easily inject and retrieve config via `context.Context` or `*http.Request`.

## Installation
```bash
go get github.com/kittipat1413/go-common
```

## Usage
This is the simplest way to get started using `MustConfig` with some defaults:
```go
package main

import (
	"fmt"

	"github.com/kittipat1413/go-common/framework/config"
)

func main() {
	cfg := config.MustConfig(
        config.WithRequiredConfigPath("env.yaml"),
		config.WithDefaults(map[string]any{
			"SERVICE_NAME": "my-service",
			"SERVICE_PORT": ":8080",
			"ENV":          "development",
		}),
	)

	// Read config values
	serviceName := cfg.GetString("SERVICE_NAME")
	port := cfg.GetString("SERVICE_PORT")
	env := cfg.GetString("ENV")

	fmt.Println("=== Service Config ===")
	fmt.Printf("Service Name: %s\n", serviceName)
	fmt.Printf("Port:         %s\n", port)
	fmt.Printf("Environment:  %s\n", env)
}
```
With an optional `env.yaml` override:
```yaml
SERVICE_NAME: "user-api"
SERVICE_PORT: ":9090"
ENV: "staging"
```

### Examples
- You can find a complete working example in the repository under [framework/config/example](example/).

## Functional Options
`WithRequiredConfigPath(path string)`: Fails if the file does not exist or is unreadable.
```go
config.WithRequiredConfigPath("env.yaml")
```

`WithOptionalConfigPaths(path string)`: Tries each path in order and uses the first found file. Skips missing files.
```go
config.WithOptionalConfigPaths("env.yaml")
```

`WithDefaults(defaults map[string]any)`: Injects fallback values if the config key is not set in env or file.
```go
config.WithDefaults(map[string]any{
    "SERVICE_NAME": "my-service",
    "SERVICE_PORT": ":8080",
    "ENV":          "development",
})
```

## Accessing Values
```go
cfg.Get("SERVICE_NAME")           // any type
cfg.GetString("SERVICE_NAME")     // string
cfg.GetInt("MAX_WORKERS")         // int
cfg.GetDuration("TIMEOUT")        // string, use time.ParseDuration
cfg.All()                         // map[string]interface{}
```