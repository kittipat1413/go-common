package sftp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kittipat1413/go-common/framework/retry"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//go:generate mockgen -source=./connection.go -destination=./mocks/connection.go -package=sftp_mocks

// ConnectionManager handles SSH/SFTP connection lifecycle
type ConnectionManager interface {
	GetConnection(ctx context.Context) (*sftp.Client, error)
	ReleaseConnection(client *sftp.Client) error
	Close() error
}

// pooledConnection represents a connection in the pool
type pooledConnection struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	lastUsed   time.Time
	inUse      bool
}

// connectionPool manages a pool of SFTP connections
type connectionPool struct {
	authConfig       AuthConfig
	authHandler      AuthenticationHandler
	connectionConfig ConnectionConfig
	connections      []*pooledConnection
	closed           bool
	retrier          retry.Retrier

	// Protects access to connections and closed flag
	mutex sync.RWMutex
	// Channel to signal cleanup routine to stop
	cleanupDone chan struct{}
	cleanupOnce sync.Once
}

// NewConnectionManager creates a new connection manager with pooling support
func NewConnectionManager(authHandler AuthenticationHandler, authConfig AuthConfig, connectionConfig ConnectionConfig) (ConnectionManager, error) {
	// Merge with default config to ensure all fields are set
	mergedAuthConfig := mergeAuthConfig(DefaultConfig().Authentication, authConfig)
	if err := validateAuthConfig(mergedAuthConfig); err != nil {
		return nil, err // errors are wrapped in validateConfig
	}
	mergedConnectionConfig := mergeConnectionConfig(DefaultConfig().Connection, connectionConfig)
	if err := validateConnectionConfig(mergedConnectionConfig); err != nil {
		return nil, err // errors are wrapped in validateConfig
	}

	if authHandler == nil {
		return nil, fmt.Errorf("%w: authentication handler cannot be nil", ErrConfiguration)
	}

	retrier, err := retry.NewRetrier(connectionConfig.RetryPolicy)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create retrier: %v", ErrConfiguration, err)
	}

	// Initialize connection pool
	cp := &connectionPool{
		authConfig:       mergedAuthConfig,
		authHandler:      authHandler,
		connectionConfig: mergedConnectionConfig,
		connections:      make([]*pooledConnection, 0, mergedConnectionConfig.MaxConnections),
		closed:           false,
		retrier:          retrier,
		cleanupDone:      make(chan struct{}),
	}

	// Auto-start cleanup routine if idle timeout is configured
	if cp.connectionConfig.IdleTimeout > 0 {
		go cp.startCleanupRoutine()
	}

	return cp, nil
}

// GetConnection retrieves or creates an SFTP connection from the pool
func (cp *connectionPool) GetConnection(ctx context.Context) (*sftp.Client, error) {
	var sftpClient *sftp.Client

	// Attempt to get or create a connection with retry logic
	err := cp.retrier.ExecuteWithRetry(ctx, func(ctx context.Context) error {
		cp.mutex.Lock()
		defer cp.mutex.Unlock()

		// Check if pool is closed
		if cp.closed {
			return fmt.Errorf("%w: connection pool is closed", ErrConnectionClosed)
		}

		// Try to find an available connection
		for i, conn := range cp.connections {
			if !conn.inUse {
				if cp.isConnectionHealthy(conn) {
					conn.inUse = true
					conn.lastUsed = time.Now()
					sftpClient = conn.sftpClient
					return nil
				} else {
					// Remove unhealthy connection
					cp.removeConnectionAtIndex(i)
					break
				}
			}
		}

		// If no available connection and we haven't reached the limit, create a new one
		if len(cp.connections) < cp.connectionConfig.MaxConnections {
			conn, err := cp.createConnectionWithRetry(ctx)
			if err != nil {
				return err // errors are wrapped in createConnectionWithRetry
			}
			cp.connections = append(cp.connections, conn)
			sftpClient = conn.sftpClient
			return nil
		}

		// Pool is full, return error
		return fmt.Errorf("%w: no available connections in the pool", ErrConnectionPoolFull)
	}, func(attempt int, err error) bool {
		// Do not retry on connection closed or authentication errors
		return !errors.Is(err, ErrConnectionClosed) &&
			!errors.Is(err, ErrAuthentication)
	})

	if err != nil {
		return nil, err // errors are wrapped in ExecuteWithRetry
	}

	if sftpClient == nil {
		return nil, fmt.Errorf("%w: unexpected error retrieving connection", ErrConnection)
	}
	return sftpClient, nil
}

// ReleaseConnection returns a connection to the pool
func (cp *connectionPool) ReleaseConnection(client *sftp.Client) error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	for _, conn := range cp.connections {
		if conn.sftpClient == client {
			conn.inUse = false
			conn.lastUsed = time.Now()
			return nil
		}
	}

	return fmt.Errorf("%w: connection not found in pool", ErrConnectionNotFound)
}

// Close closes all connections in the pool
func (cp *connectionPool) Close() error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if cp.closed {
		return nil
	}

	// Mark pool as closed
	cp.closed = true

	// Stop cleanup routine
	cp.cleanupOnce.Do(func() {
		close(cp.cleanupDone)
	})

	// Close all connections
	var lastErr error
	for _, conn := range cp.connections {
		if err := cp.closeConnection(conn); err != nil {
			lastErr = err
		}
	}

	cp.connections = nil
	return lastErr
}

// createConnectionWithRetry establishes a new SSH/SFTP connection with retry logic
func (cp *connectionPool) createConnectionWithRetry(ctx context.Context) (*pooledConnection, error) {
	var pooledConnection *pooledConnection
	err := cp.retrier.ExecuteWithRetry(ctx, func(ctx context.Context) error {
		// Create new connection
		connection, err := cp.createConnection(ctx)
		if err != nil {
			return err // errors are wrapped in createConnection
		}

		// Successfully created connection
		pooledConnection = connection
		return nil
	}, func(attempt int, err error) bool {
		// Do not retry on authentication errors
		return !errors.Is(err, ErrAuthentication)
	})

	if err != nil {
		return nil, err // errors are wrapped in ExecuteWithRetry
	}

	if pooledConnection != nil {
		return pooledConnection, nil
	}

	return nil, fmt.Errorf("%w: unexpected result type from connection creation", ErrConnection)
}

// createConnection establishes a new SSH/SFTP connection
func (cp *connectionPool) createConnection(ctx context.Context) (*pooledConnection, error) {
	// Get authentication methods
	authMethods, err := cp.authHandler.GetAuthMethods()
	if err != nil {
		return nil, err // errors are wrapped in GetAuthMethods
	}

	// Create SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            cp.authConfig.Username,
		Auth:            authMethods,
		HostKeyCallback: cp.getHostKeyCallback(),
		Timeout:         cp.connectionConfig.Timeout,
	}

	// Establish SSH connection
	address := fmt.Sprintf("%s:%d", cp.authConfig.Host, cp.authConfig.Port)
	var sshClient *ssh.Client
	if cp.connectionConfig.Timeout > 0 {
		// Use context with timeout for connection establishment
		conn, err := cp.dialWithTimeout(ctx, "tcp", address, cp.connectionConfig.Timeout)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to dial SSH: %v", ErrConnection, err)
		}

		sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("%w: SSH client connection failed: %v", ErrConnection, err)
		}

		sshClient = ssh.NewClient(sshConn, chans, reqs)
	} else {
		var err error
		sshClient, err = ssh.Dial("tcp", address, sshConfig)
		if err != nil {
			return nil, fmt.Errorf("%w: SSH dial failed: %v", ErrConnection, err)
		}
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("%w: failed to create SFTP client: %v", ErrConnection, err)
	}

	return &pooledConnection{
		sshClient:  sshClient,
		sftpClient: sftpClient,
		lastUsed:   time.Now(),
		inUse:      true,
	}, nil
}

// dialWithTimeout creates a network connection with timeout
func (cp *connectionPool) dialWithTimeout(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return dialer.DialContext(ctx, network, address)
}

// getHostKeyCallback returns the appropriate host key callback
func (cp *connectionPool) getHostKeyCallback() ssh.HostKeyCallback {
	if cp.authConfig.HostKeyCallback != nil {
		return cp.authConfig.HostKeyCallback
	}
	// Default to insecure callback (should be configurable in production)
	return ssh.InsecureIgnoreHostKey() // #nosec G106
}

// isConnectionHealthy checks if a connection is still healthy
func (cp *connectionPool) isConnectionHealthy(conn *pooledConnection) bool {
	if conn.sshClient == nil || conn.sftpClient == nil {
		return false
	}

	// Check if connection has been idle too long
	if cp.connectionConfig.IdleTimeout > 0 {
		if time.Since(conn.lastUsed) > cp.connectionConfig.IdleTimeout {
			_ = cp.closeConnection(conn)
			return false
		}
	}

	// Simple health check - try to get working directory
	_, err := conn.sftpClient.Getwd()
	return err == nil
}

// closeConnection closes a single pooled connection
func (cp *connectionPool) closeConnection(conn *pooledConnection) error {
	var lastErr error

	if conn.sftpClient != nil {
		if err := conn.sftpClient.Close(); err != nil {
			lastErr = err
		}
		conn.sftpClient = nil
	}

	if conn.sshClient != nil {
		if err := conn.sshClient.Close(); err != nil {
			lastErr = err
		}
		conn.sshClient = nil
	}

	return lastErr
}

// cleanupIdleConnections removes idle connections from the pool
func (cp *connectionPool) cleanupIdleConnections() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if cp.connectionConfig.IdleTimeout <= 0 {
		return
	}

	activeConnections := make([]*pooledConnection, 0, len(cp.connections))

	for _, conn := range cp.connections {
		if conn.inUse || time.Since(conn.lastUsed) <= cp.connectionConfig.IdleTimeout {
			activeConnections = append(activeConnections, conn)
		} else {
			_ = cp.closeConnection(conn)
		}
	}

	cp.connections = activeConnections
}

// removeConnectionAtIndex removes a connection at the specified index
func (cp *connectionPool) removeConnectionAtIndex(index int) {
	if index < 0 || index >= len(cp.connections) {
		return
	}

	conn := cp.connections[index]
	_ = cp.closeConnection(conn)

	// Remove from slice
	cp.connections = append(cp.connections[:index], cp.connections[index+1:]...)
}

// startCleanupRoutine starts a background routine to clean up idle connections
// This is called automatically if idle timeout is configured
func (cp *connectionPool) startCleanupRoutine() {
	idleTimeout := cp.connectionConfig.IdleTimeout
	if idleTimeout <= 0 {
		return
	}

	// Determine cleanup interval
	const minInterval = 1 * time.Second
	const maxInterval = 1 * time.Minute

	cleanupInterval := idleTimeout / 2
	if cleanupInterval < minInterval {
		cleanupInterval = minInterval
	}
	if cleanupInterval > maxInterval {
		cleanupInterval = maxInterval
	}

	// Start cleanup ticker
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cp.cleanupDone:
			return
		case <-ticker.C:
			cp.cleanupIdleConnections()
		}
	}
}
