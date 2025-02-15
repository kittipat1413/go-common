[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Total Views](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fkittipat1413%2Fgo-common&count_bg=%2379C83D&title_bg=%23555555&icon=go.svg&icon_color=%23E7E7E7&title=Total+Views&edge_flat=false)](https://hits.seeyoufarm.com)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# 🔄 Retry Package
The Retry Package provides a robust and extensible interface for automatically retrying operations in Go. It supports configurable retry strategies like fixed delays, jitter, and exponential backoff, ensuring reliability in API calls, database queries, and distributed systems.

## Features
- **Customizable Backoff Strategies** – Supports Fixed, Jitter, and Exponential backoff
- **Context-Aware** – Automatically stops retries when the context is canceled
- **Configurable Retry Conditions** – Choose which errors should trigger retries

## Usage
### Retrier Interface
```go
func (r *Retrier) ExecuteWithRetry(ctx context.Context, fn RetryFunc, retryOn RetryOnFunc) error
```
**ExecuteWithRetry**: Executes a function with automatic retry logic.
- **Params**:
    - `ctx`: Context for request tracing and cancellation
    - `fn`: The function to retry (must return an error if it fails)
    - `retryOn`: Custom function to determine retry conditions
- **Returns**: 
    - `error`: The final result after retries.

### Example: Basic Retry
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kittipat1413/go-common/framework/retry"
)

func main() {
	ctx := context.Background()

	// Define retry configuration
	config := retry.Config{
		MaxAttempts: 3,
		Backoff:     &retry.FixedBackoff{Interval: 2 * time.Second},
	}

	// Create Retrier
	retrier, err := retry.NewRetrier(config)
	if err != nil {
		log.Fatalf("Failed to create retrier: %v", err)
	}

	// Execute function with retry logic
	err = retrier.ExecuteWithRetry(ctx, func(ctx context.Context) error {
		fmt.Println("Attempting API request...")
		return errors.New("network timeout")
	}, func(attempt int, err error) bool {
		fmt.Printf("Retry %d due to: %v\n", attempt, err)
		return err.Error() == "network timeout" // Retry only for network timeouts
	})

	if err != nil {
		fmt.Println("Final failure:", err)
	} else {
		fmt.Println("Operation succeeded!")
	}
}
```
You can find a complete working example in the repository under [framework/retry/example](example/).


## Backoff Strategies
**1. Fixed Backoff**
```go
backoff := &retry.FixedBackoff{Interval: 2 * time.Second}
```
- Constant delay between retries 
- Simple and predictable retry behavior

**2. Jitter Backoff**
```go
backoff := &retry.JitterBackoff{BaseDelay: 2 * time.Second, MaxJitter: 500 * time.Millisecond}
```
- Adds randomness to prevent synchronized retries (thundering herd problem)

**3. Exponential Backoff**
```go
backoff := &retry.ExponentialBackoff{BaseDelay: 100 * time.Millisecond, Factor: 2.0, MaxDelay: 5 * time.Second}
```
- Delays grow exponentially (`BaseDelay` * `Factor`^`attempt`)
- Prevents excessive load on failing services