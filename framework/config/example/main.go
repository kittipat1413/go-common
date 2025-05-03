package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kittipat1413/go-common/framework/config"
)

// DatabaseConfig holds the database configuration values.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Environment variable keys for database config
const (
	databaseUrlKey             = "DATABASE_URL"
	databaseMaxOpenConnsKey    = "DATABASE_MAX_OPEN_CONNS"
	databaseMaxIdleConnsKey    = "DATABASE_MAX_IDLE_CONNS"
	databaseConnMaxLifetimeKey = "DATABASE_CONN_MAX_LIFETIME"
	databaseConnMaxIdleTimeKey = "DATABASE_CONN_MAX_IDLE_TIME"
)

func main() {
	cfg := config.MustConfig(
		config.WithOptionalConfigPaths("env.yaml", "../env.yaml", "framework/config/example/env.yaml"),
		config.WithDefaults(map[string]any{
			databaseUrlKey:             "postgres://localhost:5432/mydb?sslmode=disable",
			databaseMaxOpenConnsKey:    20,
			databaseMaxIdleConnsKey:    10,
			databaseConnMaxLifetimeKey: "30m",
			databaseConnMaxIdleTimeKey: "5m",
		}),
	)

	dbCfg, err := loadDatabaseConfig(cfg)
	if err != nil {
		log.Fatalf("failed to load database config: %v", err)
	}

	fmt.Println("Database config:")
	fmt.Printf("  URL:               %s\n", dbCfg.URL)
	fmt.Printf("  MaxOpenConns:      %d\n", dbCfg.MaxOpenConns)
	fmt.Printf("  MaxIdleConns:      %d\n", dbCfg.MaxIdleConns)
	fmt.Printf("  ConnMaxLifetime:   %v\n", dbCfg.ConnMaxLifetime)
	fmt.Printf("  ConnMaxIdleTime:   %v\n", dbCfg.ConnMaxIdleTime)
}

// loadDatabaseConfig reads and parses database config values from the provided config loader.
func loadDatabaseConfig(cfg *config.Config) (*DatabaseConfig, error) {
	lifetime, err := time.ParseDuration(cfg.GetString(databaseConnMaxLifetimeKey))
	if err != nil {
		return nil, fmt.Errorf("invalid duration for %s: %w", databaseConnMaxLifetimeKey, err)
	}

	idleTime, err := time.ParseDuration(cfg.GetString(databaseConnMaxIdleTimeKey))
	if err != nil {
		return nil, fmt.Errorf("invalid duration for %s: %w", databaseConnMaxIdleTimeKey, err)
	}

	return &DatabaseConfig{
		URL:             cfg.GetString(databaseUrlKey),
		MaxOpenConns:    cfg.GetInt(databaseMaxOpenConnsKey),
		MaxIdleConns:    cfg.GetInt(databaseMaxIdleConnsKey),
		ConnMaxLifetime: lifetime,
		ConnMaxIdleTime: idleTime,
	}, nil
}
