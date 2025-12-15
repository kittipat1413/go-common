package sftp

import (
	"errors"
)

// Error definitions for configuration issues
var (
	ErrConfiguration = errors.New("configuration error")
)

// Error definitions for authentication issues
var (
	ErrAuthentication = errors.New("authentication error")
)

// Error definitions for SFTP connection management
var (
	ErrConnection         = errors.New("connection error")
	ErrConnectionPoolFull = errors.New("connection pool full")
	ErrConnectionClosed   = errors.New("connection closed")
	ErrConnectionNotFound = errors.New("connection not found")
)

// Error definitions for file operations
var (
	ErrFileNotFound = errors.New("file not found")
	ErrDataTransfer = errors.New("data transfer error")
)
