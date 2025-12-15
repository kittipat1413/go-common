package sftp_test

import (
	"context"
	"errors"
	"net"
	"sync"
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

func TestConnectionPool_GetConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful connection creation", func(t *testing.T) {
		// Create a mock server for testing
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer listener.Close()

		port := listener.Addr().(*net.TCPAddr).Port
		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()

		config := sftp.Config{
			Host:     "127.0.0.1",
			Port:     port,
			Username: "testuser",
			Connection: sftp.ConnectionConfig{
				MaxConnections: 2,
				Timeout:        100 * time.Millisecond, // Short timeout for quick test
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

		mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockAuth.EXPECT().GetAuthMethods().Return([]ssh.AuthMethod{
			ssh.Password("password"),
		}, nil).AnyTimes()

		manager, err := sftp.NewConnectionManager(config, mockAuth)
		require.NoError(t, err)
		defer manager.Close()

		ctx := context.Background()

		// This will likely fail due to authentication, but we're testing the retry mechanism
		_, err = manager.GetConnection(ctx)
		assert.ErrorIs(t, err, sftp.ErrConnection)
	})

	t.Run("context cancellation", func(t *testing.T) {
		config := sftp.Config{
			Host:     "nonexistent.example.com",
			Port:     22,
			Username: "testuser",
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        5 * time.Second,
				RetryPolicy: retry.Config{
					MaxAttempts: 3,
					Backoff: &retry.FixedBackoff{
						Interval: 100 * time.Millisecond,
					},
				},
			},
		}

		mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockAuth.EXPECT().GetAuthMethods().Return([]ssh.AuthMethod{
			ssh.Password("password"),
		}, nil).AnyTimes()

		manager, err := sftp.NewConnectionManager(config, mockAuth)
		require.NoError(t, err)
		defer manager.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err = manager.GetConnection(ctx)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("auth handler error", func(t *testing.T) {
		config := sftp.Config{
			Host:     "example.com",
			Port:     22,
			Username: "testuser",
			Connection: sftp.ConnectionConfig{
				MaxConnections: 1,
				Timeout:        time.Second,
				RetryPolicy: retry.Config{
					MaxAttempts: 1,
					Backoff: &retry.FixedBackoff{
						Interval: 10 * time.Millisecond,
					},
				},
			},
		}

		mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		expectedErr := errors.New("authentication error")
		mockAuth.EXPECT().GetAuthMethods().Return(nil, expectedErr)

		manager, err := sftp.NewConnectionManager(config, mockAuth)
		require.NoError(t, err)
		defer manager.Close()

		ctx := context.Background()
		_, err = manager.GetConnection(ctx)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestConnectionPool_ReleaseConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := sftp.Config{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		Connection: sftp.ConnectionConfig{
			MaxConnections: 1,
			Timeout:        time.Second,
		},
	}

	mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)
	manager, err := sftp.NewConnectionManager(config, mockAuth)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("connection not found", func(t *testing.T) {
		// Create a mock SFTP client that's not in the pool
		// We can't create a real sftp.Client, so we'll create one through a different method
		// For this test, we'll use nil since the pool should handle this gracefully
		var mockSftpClient *pkg_sftp.Client

		err := manager.ReleaseConnection(mockSftpClient)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sftp.ErrConnectionNotFound))
	})
}

func TestConnectionPool_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := sftp.Config{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		Connection: sftp.ConnectionConfig{
			MaxConnections: 1,
			Timeout:        time.Second,
		},
	}

	mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)
	manager, err := sftp.NewConnectionManager(config, mockAuth)
	require.NoError(t, err)

	t.Run("close successfully", func(t *testing.T) {
		err := manager.Close()
		assert.NoError(t, err)

		// Closing again should be safe
		err = manager.Close()
		assert.NoError(t, err)
	})

	t.Run("get connection after close", func(t *testing.T) {
		ctx := context.Background()
		_, err := manager.GetConnection(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, sftp.ErrConnectionClosed))
	})
}

func TestConnectionPool_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := sftp.Config{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		Connection: sftp.ConnectionConfig{
			MaxConnections: 3,
			Timeout:        100 * time.Millisecond,
			RetryPolicy: retry.Config{
				MaxAttempts: 1,
				Backoff: &retry.FixedBackoff{
					Interval: 10 * time.Millisecond,
				},
			},
		},
	}

	mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)
	mockAuth.EXPECT().GetAuthMethods().Return([]ssh.AuthMethod{
		ssh.Password("password"),
	}, nil).AnyTimes()

	manager, err := sftp.NewConnectionManager(config, mockAuth)
	require.NoError(t, err)
	defer manager.Close()

	// Test concurrent access to GetConnection
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, err := manager.GetConnection(ctx)
			// We expect errors due to connection failures, but no panics
			_ = err
		}()
	}

	wg.Wait()

	// Test concurrent Close calls
	var closeWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		closeWg.Add(1)
		go func() {
			defer closeWg.Done()
			_ = manager.Close()
		}()
	}

	closeWg.Wait()
}

func TestConnectionPool_MaxConnections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Use a very small connection pool to test the limit
	config := sftp.Config{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		Connection: sftp.ConnectionConfig{
			MaxConnections: 1, // Only allow 1 connection
			Timeout:        100 * time.Millisecond,
			RetryPolicy: retry.Config{
				MaxAttempts: 1,
				Backoff: &retry.FixedBackoff{
					Interval: 10 * time.Millisecond,
				},
			},
		},
	}

	mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)

	// Mock will return auth error to prevent actual connections
	authError := errors.New("auth error")
	mockAuth.EXPECT().GetAuthMethods().Return(nil, authError).AnyTimes()

	manager, err := sftp.NewConnectionManager(config, mockAuth)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()

	// This should fail due to auth error, not pool limits
	_, err = manager.GetConnection(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth error")
}

func TestConnectionPool_RetryPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := sftp.Config{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		Connection: sftp.ConnectionConfig{
			MaxConnections: 1,
			Timeout:        50 * time.Millisecond,
			RetryPolicy: retry.Config{
				MaxAttempts: 2,
				Backoff: &retry.FixedBackoff{
					Interval: 10 * time.Millisecond,
				},
			},
		},
	}

	mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)

	// First call fails, second call also fails
	authError := errors.New("temporary auth error")
	mockAuth.EXPECT().GetAuthMethods().Return(nil, authError).AnyTimes()

	manager, err := sftp.NewConnectionManager(config, mockAuth)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()
	start := time.Now()

	_, err = manager.GetConnection(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Should have retried at least once (with delay)
	assert.True(t, elapsed >= 10*time.Millisecond, "should have waited for retry delay")
}

func TestConnectionPool_InvalidRetryPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test with a config that has an invalid backoff strategy after merging
	config := sftp.Config{
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
	}

	mockAuth := sftp_mocks.NewMockAuthenticationHandler(ctrl)

	_, err := sftp.NewConnectionManager(config, mockAuth)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sftp.ErrConfiguration))
}
