package sftp_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kittipat1413/go-common/framework/retry"
	"github.com/kittipat1413/go-common/framework/sftp"
	sftp_mocks "github.com/kittipat1413/go-common/framework/sftp/mocks"
	pkg_sftp "github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// testSFTPServer represents an in-memory SFTP server for testing
type testSFTPServer struct {
	listener     net.Listener
	serverConfig *ssh.ServerConfig
	tempDir      string
	auth         testSFTPServerAuth
	wg           sync.WaitGroup
	mutex        sync.Mutex
	closed       bool
}

type testSFTPServerAuth struct {
	username      string
	password      string
	hostKey       ssh.Signer
	publicKey     ssh.PublicKey
	privateKeyPEM []byte
}

// newTestSFTPServer creates a new in-memory SFTP server
func newTestSFTPServer(t *testing.T) *testSFTPServer {
	t.Helper()

	// Create temporary directory for file operations
	tempDir, err := os.MkdirTemp("", "sftp-test-*")
	require.NoError(t, err)

	// Generate RSA key pair for the server
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create SSH signer from private key
	signer, err := ssh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	// Extract public key for client authentication
	publicKey := signer.PublicKey()

	// Create PEM-encoded private key for client use
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	// Setup authentication credentials
	auth := testSFTPServerAuth{
		username:      "testuser",
		password:      "testpass",
		hostKey:       signer,
		publicKey:     publicKey,
		privateKeyPEM: privateKeyPEM,
	}

	// Configure SSH server
	config := &ssh.ServerConfig{
		// Accept any password for testing
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == auth.username && string(pass) == auth.password {
				return &ssh.Permissions{}, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		// Accept our test public key
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if c.User() != auth.username {
				return nil, fmt.Errorf("public key rejected for %q", c.User())
			}
			if auth.publicKey != nil &&
				bytes.Equal(pubKey.Marshal(), auth.publicKey.Marshal()) {
				return &ssh.Permissions{}, nil
			}
			return nil, fmt.Errorf("public key mismatch for %q", c.User())
		},
		NoClientAuth: false,
	}
	config.AddHostKey(signer)

	// Create listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &testSFTPServer{
		listener:     listener,
		serverConfig: config,
		tempDir:      tempDir,
		auth:         auth,
	}

	// Start accepting connections
	server.wg.Add(1)
	go server.serve()

	// Ensure server is closed on test cleanup
	t.Cleanup(func() {
		_ = server.close()
	})

	return server
}

// serve handles incoming connections
func (s *testSFTPServer) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mutex.Lock()
			closed := s.closed
			s.mutex.Unlock()
			if closed {
				return
			}
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single SSH connection
func (s *testSFTPServer) handleConnection(netConn net.Conn) {
	defer s.wg.Done()
	defer netConn.Close()

	// Set a reasonable timeout for the connection
	_ = netConn.SetDeadline(time.Now().Add(30 * time.Second))

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.serverConfig)
	if err != nil {
		return
	}
	defer sshConn.Close()

	// Handle global SSH requests
	go ssh.DiscardRequests(reqs)

	// Handle SSH channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		s.wg.Add(1)
		go s.handleChannel(channel, requests)
	}
}

// handleChannel handles SSH channel requests (SFTP subsystem)
func (s *testSFTPServer) handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer s.wg.Done()
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "subsystem":
			// Extract subsystem name from payload
			if len(req.Payload) < 4 {
				_ = req.Reply(false, nil)
				continue
			}

			// SSH string format: 4 bytes length + data
			subsysLen := uint32(req.Payload[0])<<24 | uint32(req.Payload[1])<<16 | uint32(req.Payload[2])<<8 | uint32(req.Payload[3])
			if subsysLen > uint32(len(req.Payload)-4) {
				_ = req.Reply(false, nil)
				continue
			}

			subsysName := string(req.Payload[4 : 4+subsysLen])

			if subsysName == "sftp" {
				_ = req.Reply(true, nil)
				// Start SFTP server
				server, err := pkg_sftp.NewServer(channel, pkg_sftp.WithServerWorkingDirectory(s.tempDir))
				if err != nil {
					return
				}
				_ = server.Serve()
				return
			}
			_ = req.Reply(false, nil)
		default:
			_ = req.Reply(false, nil)
		}
	}
}

// getAddress returns the server's listening address
func (s *testSFTPServer) getAddress() string {
	return s.listener.Addr().(*net.TCPAddr).IP.String()
}

// getPort returns the server's listening port
func (s *testSFTPServer) getPort() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// close stops the server and cleans up
func (s *testSFTPServer) close() error {
	s.mutex.Lock()
	s.closed = true
	s.mutex.Unlock()

	err := s.listener.Close()
	s.wg.Wait()

	// Clean up temp directory
	os.RemoveAll(s.tempDir)
	return err
}

func TestNewConnectionManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		config      sftp.Config
		authHandler sftp.AuthenticationHandler
		expectError bool
		errorType   error
	}{
		{
			name: "valid config and auth handler",
			config: sftp.Config{
				Host:     "example.com",
				Port:     22,
				Username: "testuser",
				Connection: sftp.ConnectionConfig{
					MaxConnections: 5,
					Timeout:        30 * time.Second,
					IdleTimeout:    5 * time.Minute,
					RetryPolicy: retry.Config{
						MaxAttempts: 3,
						Backoff: &retry.ExponentialBackoff{
							BaseDelay: time.Second,
							Factor:    2.0,
							MaxDelay:  30 * time.Second,
						},
					},
				},
			},
			authHandler: sftp_mocks.NewMockAuthenticationHandler(ctrl),
			expectError: false,
		},
		{
			name: "nil auth handler",
			config: sftp.Config{
				Host:     "example.com",
				Port:     22,
				Username: "testuser",
			},
			authHandler: nil,
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - empty host",
			config: sftp.Config{
				Port:     22,
				Username: "testuser",
			},
			authHandler: sftp_mocks.NewMockAuthenticationHandler(ctrl),
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - negative port",
			config: sftp.Config{
				Host:     "example.com",
				Port:     -1,
				Username: "testuser",
			},
			authHandler: sftp_mocks.NewMockAuthenticationHandler(ctrl),
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - empty username",
			config: sftp.Config{
				Host: "example.com",
				Port: 22,
			},
			authHandler: sftp_mocks.NewMockAuthenticationHandler(ctrl),
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - invalid retry policy",
			config: sftp.Config{
				Host:     "example.com",
				Port:     22,
				Username: "testuser",
				Connection: sftp.ConnectionConfig{
					MaxConnections: 1,
					Timeout:        time.Second,
					RetryPolicy: retry.Config{
						MaxAttempts: 1,
						Backoff: &retry.FixedBackoff{
							Interval: -1 * time.Second, // Invalid - negative interval
						},
					},
				},
			},
			authHandler: sftp_mocks.NewMockAuthenticationHandler(ctrl),
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := sftp.NewConnectionManager(tt.config, tt.authHandler)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
				if tt.errorType != nil {
					assert.True(t, errors.Is(err, tt.errorType))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)

				// Clean up
				if manager != nil {
					_ = manager.Close()
				}
			}
		})
	}
}

func TestConnectionPool(t *testing.T) {
	server := newTestSFTPServer(t)
	defer server.close()

	baseConfig := sftp.Config{
		Host:     server.getAddress(),
		Port:     server.getPort(),
		Username: server.auth.username,
		Connection: sftp.ConnectionConfig{
			MaxConnections: 3,
			Timeout:        5 * time.Second,
			IdleTimeout:    30 * time.Second,
			RetryPolicy: retry.Config{
				MaxAttempts: 3,
				Backoff: &retry.ExponentialBackoff{
					BaseDelay: 100 * time.Millisecond,
					Factor:    2.0,
					MaxDelay:  1 * time.Second,
				},
			},
		},
		Authentication: sftp.AuthConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	t.Run("Password Authentication", func(t *testing.T) {
		config := baseConfig
		config.Authentication.Method = sftp.AuthPassword
		config.Authentication.Password = server.auth.password

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Test getting a connection
		client, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Test that the connection works
		workDir, err := client.Getwd()
		require.NoError(t, err)
		require.NotEmpty(t, workDir)

		// Release connection
		require.NoError(t, pool.ReleaseConnection(client))
	})

	t.Run("Private Key Authentication", func(t *testing.T) {
		config := baseConfig
		config.Authentication.Method = sftp.AuthPrivateKey
		config.Authentication.PrivateKeyData = server.auth.privateKeyPEM

		authHandler := sftp.NewPrivateKeyAuthHandlerWithData(server.auth.username, server.auth.privateKeyPEM, "")
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Test getting a connection
		client, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Test that the connection works
		_, err = client.Getwd()
		require.NoError(t, err)

		// Release connection
		require.NoError(t, pool.ReleaseConnection(client))
	})

	t.Run("Connection Pool Behavior", func(t *testing.T) {
		config := baseConfig
		config.Connection.MaxConnections = 2
		config.Authentication.Method = sftp.AuthPassword
		config.Authentication.Password = server.auth.password

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Get first connection
		client1, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client1)

		// Get second connection
		client2, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client2)

		// Both connections should work
		_, err = client1.Getwd()
		require.NoError(t, err)

		_, err = client2.Getwd()
		require.NoError(t, err)

		// Try to get third connection - should block and timeout
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		_, err = pool.GetConnection(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")

		// Release first connection
		require.NoError(t, pool.ReleaseConnection(client1))

		// Now we should be able to get a connection again
		ctx = context.Background()
		client3, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client3)

		// Clean up
		require.NoError(t, pool.ReleaseConnection(client2))
		require.NoError(t, pool.ReleaseConnection(client3))
	})

	t.Run("Connection Reuse", func(t *testing.T) {
		config := baseConfig
		config.Connection.MaxConnections = 1
		config.Authentication.Method = sftp.AuthPassword
		config.Authentication.Password = server.auth.password

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Get a connection
		client1, err := pool.GetConnection(ctx)
		require.NoError(t, err)

		// Test it works
		_, err = client1.Getwd()
		require.NoError(t, err)

		// Release it
		require.NoError(t, pool.ReleaseConnection(client1))

		// Get another connection - should reuse the first one
		client2, err := pool.GetConnection(ctx)
		require.NoError(t, err)

		// Should still work
		_, err = client2.Getwd()
		require.NoError(t, err)

		require.NoError(t, pool.ReleaseConnection(client2))
	})

	t.Run("Idle Connection Cleanup", func(t *testing.T) {
		config := baseConfig
		config.Connection.MaxConnections = 2
		config.Connection.IdleTimeout = 100 * time.Millisecond // Very short for testing
		config.Authentication.Method = sftp.AuthPassword
		config.Authentication.Password = server.auth.password

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Get a connection
		client, err := pool.GetConnection(ctx)
		require.NoError(t, err)

		// Test it works
		_, err = client.Getwd()
		require.NoError(t, err)

		// Release it
		require.NoError(t, pool.ReleaseConnection(client))

		// Wait for cleanup to happen
		time.Sleep(1 * time.Second)

		// Get another connection - the old one should have been cleaned up
		// and a new one created
		client2, err := pool.GetConnection(ctx)
		require.NoError(t, err)

		// Should still work
		_, err = client2.Getwd()
		require.NoError(t, err)

		require.NoError(t, pool.ReleaseConnection(client2))
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		config := baseConfig
		config.Connection.MaxConnections = 5
		config.Authentication.Method = sftp.AuthPassword
		config.Authentication.Password = server.auth.password

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		var wg sync.WaitGroup
		numGoroutines := 10
		var successCount int32

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				ctx := context.Background()
				client, err := pool.GetConnection(ctx)
				if err != nil {
					return
				}

				// Test the connection
				if _, err := client.Getwd(); err != nil {
					return
				}

				// Release connection
				if err := pool.ReleaseConnection(client); err == nil {
					atomic.AddInt32(&successCount, 1)
				}
			}(i)
		}

		wg.Wait()

		// At least some operations should have succeeded
		assert.True(t, atomic.LoadInt32(&successCount) > 0, "Expected at least some concurrent operations to succeed")
	})
}

func TestConnectionPool_InternalMethods(t *testing.T) {
	server := newTestSFTPServer(t)
	defer server.close()

	t.Run("Remove Connection At Invalid Index", func(t *testing.T) {
		config := sftp.Config{
			Host:     server.getAddress(),
			Port:     server.getPort(),
			Username: server.auth.username,
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        5 * time.Second,
				IdleTimeout:    100 * time.Millisecond,
				RetryPolicy: retry.Config{
					MaxAttempts: 2,
					Backoff: &retry.FixedBackoff{
						Interval: 100 * time.Millisecond,
					},
				},
			},
			Authentication: sftp.AuthConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Get a connection to populate the pool
		client1, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client1)

		// Release the connection back to the pool
		require.NoError(t, pool.ReleaseConnection(client1))

		time.Sleep(150 * time.Millisecond)

		// Get connection again - the old one should have been removed due to idle timeout
		client2, err := pool.GetConnection(ctx)
		require.NoError(t, err)
		require.NotNil(t, client2)

		_, err = client2.Getwd()
		require.NoError(t, err)

		require.NoError(t, pool.ReleaseConnection(client2))
	})

	t.Run("Release Non-existent Connection", func(t *testing.T) {
		config := sftp.Config{
			Host:     server.getAddress(),
			Port:     server.getPort(),
			Username: server.auth.username,
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        5 * time.Second,
				RetryPolicy: retry.Config{
					MaxAttempts: 2,
					Backoff: &retry.FixedBackoff{
						Interval: 100 * time.Millisecond,
					},
				},
			},
			Authentication: sftp.AuthConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()

		// Get a connection
		client1, err := pool.GetConnection(ctx)
		require.NoError(t, err)

		// Release it
		require.NoError(t, pool.ReleaseConnection(client1))

		// Try to release a nil client (not in pool)
		dummyClient := &pkg_sftp.Client{}
		err = pool.ReleaseConnection(dummyClient)
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrConnectionNotFound)
	})

	t.Run("Multiple Close Calls", func(t *testing.T) {
		config := sftp.Config{
			Host:     server.getAddress(),
			Port:     server.getPort(),
			Username: server.auth.username,
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        5 * time.Second,
				RetryPolicy: retry.Config{
					MaxAttempts: 1,
					Backoff: &retry.FixedBackoff{
						Interval: 100 * time.Millisecond,
					},
				},
			},
			Authentication: sftp.AuthConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)

		// Close the pool
		require.NoError(t, pool.Close())

		// Close again - should not error
		require.NoError(t, pool.Close())
	})
}

func TestConnectionPool_ErrorScenarios(t *testing.T) {
	server := newTestSFTPServer(t)
	defer server.close()

	t.Run("Invalid Authentication", func(t *testing.T) {
		config := sftp.Config{
			Host:     server.getAddress(),
			Port:     server.getPort(),
			Username: server.auth.username,
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        2 * time.Second,
				RetryPolicy: retry.Config{
					MaxAttempts: 2,
					Backoff: &retry.FixedBackoff{
						Interval: 100 * time.Millisecond,
					},
				},
			},
			Authentication: sftp.AuthConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, "wrongpass")
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()
		_, err = pool.GetConnection(ctx)
		require.ErrorIs(t, err, sftp.ErrConnection)
	})

	t.Run("Connection After Pool Close", func(t *testing.T) {
		config := sftp.Config{
			Host:     server.getAddress(),
			Port:     server.getPort(),
			Username: server.auth.username,
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        2 * time.Second,
				RetryPolicy: retry.Config{
					MaxAttempts: 1,
					Backoff: &retry.FixedBackoff{
						Interval: 100 * time.Millisecond,
					},
				},
			},
			Authentication: sftp.AuthConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)

		// Close the pool
		require.NoError(t, pool.Close())

		// Try to get a connection after closing
		ctx := context.Background()
		_, err = pool.GetConnection(ctx)
		require.ErrorIs(t, err, sftp.ErrConnectionClosed)
	})

	t.Run("Connection Timeout", func(t *testing.T) {
		// Test with an unreachable host to trigger timeout
		config := sftp.Config{
			Host:     "192.0.2.1", // RFC5737 test address - should be unreachable
			Port:     22,
			Username: server.auth.username,
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        100 * time.Millisecond, // Very short timeout
				RetryPolicy: retry.Config{
					MaxAttempts: 1,
					Backoff: &retry.FixedBackoff{
						Interval: 10 * time.Millisecond,
					},
				},
			},
			Authentication: sftp.AuthConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}

		authHandler := sftp.NewPasswordAuthHandler(server.auth.username, server.auth.password)
		pool, err := sftp.NewConnectionManager(config, authHandler)
		require.NoError(t, err)
		defer pool.Close()

		ctx := context.Background()
		_, err = pool.GetConnection(ctx)
		require.Error(t, err)
	})
}
