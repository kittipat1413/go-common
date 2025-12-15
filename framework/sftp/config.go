package sftp

import (
	"fmt"
	"time"

	"github.com/kittipat1413/go-common/framework/retry"
	"golang.org/x/crypto/ssh"
)

// Config contains all configuration options for the SFTP client
type Config struct {
	Host           string           // SFTP server hostname or IP address
	Port           int              // SFTP server port number
	Username       string           // Username for authentication
	Authentication AuthConfig       // Authentication-related configuration
	Connection     ConnectionConfig // Connection-related configuration
	Transfer       TransferConfig   // File transfer-related configuration
}

// AuthConfig contains authentication-related configuration
type AuthConfig struct {
	Method          AuthMethod          // Authentication method (AuthPassword or AuthPrivateKey)
	Password        string              // Password for AuthPassword method
	PrivateKeyPath  string              // Path to private key file for AuthPrivateKey method
	PrivateKeyData  []byte              // Private key data for AuthPrivateKey method
	HostKeyCallback ssh.HostKeyCallback // Host key callback for server verification
}

// ConnectionConfig contains connection-related configuration
type ConnectionConfig struct {
	Timeout        time.Duration // Connection timeout duration
	MaxConnections int           // Maximum number of simultaneous connections in the pool
	IdleTimeout    time.Duration // Idle connection timeout duration
	RetryPolicy    retry.Config  // Retry policy for connection attempts
}

// TransferConfig contains file transfer-related configuration
type TransferConfig struct {
	BufferSize          int              // Size of the buffer used during file transfers (in bytes)
	CreateDirs          bool             // Whether to create missing directories during file uploads
	PreservePermissions bool             // Whether to preserve file permissions during transfers
	ProgressCallback    ProgressCallback // Optional callback for reporting progress during file transfers
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Port: 22,
		Connection: ConnectionConfig{
			Timeout:        30 * time.Second,
			MaxConnections: 10,
			IdleTimeout:    5 * time.Minute,
			RetryPolicy: retry.Config{
				MaxAttempts: 3,
				Backoff: &retry.ExponentialBackoff{
					BaseDelay: 1 * time.Second,
					Factor:    2.0,
					MaxDelay:  30 * time.Second,
				},
			},
		},
		Transfer: TransferConfig{
			BufferSize:          32 * 1024, // 32KB
			CreateDirs:          false,
			PreservePermissions: false,
		},
	}
}

// MergeConfig merges user configuration with defaults
func MergeConfig(userConfig Config) Config {
	config := DefaultConfig()

	// Merge basic connection settings
	if userConfig.Host != "" {
		config.Host = userConfig.Host
	}
	if userConfig.Port != 0 {
		config.Port = userConfig.Port
	}
	if userConfig.Username != "" {
		config.Username = userConfig.Username
	}

	// Merge authentication config
	config.Authentication = mergeAuthConfig(config.Authentication, userConfig.Authentication)

	// Merge connection config
	config.Connection = mergeConnectionConfig(config.Connection, userConfig.Connection)

	// Merge transfer config
	config.Transfer = mergeTransferConfig(config.Transfer, userConfig.Transfer)

	return config
}

// mergeAuthConfig merges authentication configuration
func mergeAuthConfig(defaultAuth, userAuth AuthConfig) AuthConfig {
	result := defaultAuth

	if userAuth.Method != 0 {
		result.Method = userAuth.Method
	}
	if userAuth.Password != "" {
		result.Password = userAuth.Password
	}
	if userAuth.PrivateKeyPath != "" {
		result.PrivateKeyPath = userAuth.PrivateKeyPath
	}
	if len(userAuth.PrivateKeyData) > 0 {
		result.PrivateKeyData = userAuth.PrivateKeyData
	}
	if userAuth.HostKeyCallback != nil {
		result.HostKeyCallback = userAuth.HostKeyCallback
	}

	return result
}

// mergeConnectionConfig merges connection configuration
func mergeConnectionConfig(defaultConn, userConn ConnectionConfig) ConnectionConfig {
	result := defaultConn

	if userConn.Timeout > 0 {
		result.Timeout = userConn.Timeout
	}
	if userConn.MaxConnections > 0 {
		result.MaxConnections = userConn.MaxConnections
	}
	if userConn.IdleTimeout > 0 {
		result.IdleTimeout = userConn.IdleTimeout
	}

	// Merge retry policy
	result.RetryPolicy = mergeRetryPolicy(result.RetryPolicy, userConn.RetryPolicy)

	return result
}

// mergeRetryPolicy merges retry policy configuration
func mergeRetryPolicy(defaultPolicy, userPolicy retry.Config) retry.Config {
	result := defaultPolicy

	if userPolicy.MaxAttempts > 0 {
		result.MaxAttempts = userPolicy.MaxAttempts
	}
	if userPolicy.Backoff != nil {
		result.Backoff = userPolicy.Backoff
	}

	return result
}

// mergeTransferConfig merges transfer configuration
func mergeTransferConfig(defaultTransfer, userTransfer TransferConfig) TransferConfig {
	result := defaultTransfer

	if userTransfer.BufferSize > 0 {
		result.BufferSize = userTransfer.BufferSize
	}
	// Boolean fields need explicit checking since false is a valid value
	if userTransfer.CreateDirs != defaultTransfer.CreateDirs {
		result.CreateDirs = userTransfer.CreateDirs
	}
	if userTransfer.PreservePermissions != defaultTransfer.PreservePermissions {
		result.PreservePermissions = userTransfer.PreservePermissions
	}
	if userTransfer.ProgressCallback != nil {
		result.ProgressCallback = userTransfer.ProgressCallback
	}

	return result
}

// validateConfig validates the SFTP client configuration
func validateConfig(config Config) error {
	if config.Host == "" {
		return fmt.Errorf("%w: host cannot be empty", ErrConfiguration)
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535, got %d", ErrConfiguration, config.Port)
	}

	if config.Username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrConfiguration)
	}

	// Validate connection configuration
	if err := validateConnectionConfig(config.Connection); err != nil {
		return err
	}

	// Validate transfer configuration
	if err := validateTransferConfig(config.Transfer); err != nil {
		return err
	}

	return nil
}

// validateConnectionConfig validates connection-specific configuration
func validateConnectionConfig(config ConnectionConfig) error {
	if config.Timeout < 0 {
		return fmt.Errorf("%w: timeout cannot be negative", ErrConfiguration)
	}

	if config.MaxConnections <= 0 {
		return fmt.Errorf("%w: max connections must be positive, got %d", ErrConfiguration, config.MaxConnections)
	}

	if config.IdleTimeout < 0 {
		return fmt.Errorf("%w: idle timeout cannot be negative", ErrConfiguration)
	}

	// Validate retry policy
	if err := config.RetryPolicy.Validate(); err != nil {
		return fmt.Errorf("%w: invalid retry policy: %v", ErrConfiguration, err)
	}

	return nil
}

// validateTransferConfig validates transfer-specific configuration
func validateTransferConfig(config TransferConfig) error {
	if config.BufferSize <= 0 {
		return fmt.Errorf("%w: buffer size must be positive, got %d", ErrConfiguration, config.BufferSize)
	}

	// Reasonable buffer size limits
	if config.BufferSize > 10*1024*1024 { // 10MB
		return fmt.Errorf("%w: buffer size too large, got %d", ErrConfiguration, config.BufferSize)
	}

	return nil
}
