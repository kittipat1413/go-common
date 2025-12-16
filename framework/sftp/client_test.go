package sftp_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/kittipat1413/go-common/framework/sftp"
	sftp_mocks "github.com/kittipat1413/go-common/framework/sftp/mocks"
	pkg_sftp "github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      sftp.Config
		expectError bool
		errorType   error
	}{
		{
			name: "valid password authentication config",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Host:            "example.com",
					Port:            22,
					Username:        "testuser",
					Method:          sftp.AuthPassword,
					Password:        "testpass",
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				},
			},
			expectError: false,
		},
		{
			name: "valid private key authentication config with valid key",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Host:            "example.com",
					Port:            22,
					Username:        "testuser",
					Method:          sftp.AuthPrivateKey,
					PrivateKeyPath:  "/path/to/key",
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				},
			},
			expectError: false, // Config validation passes, key file validation happens at connect time
		},
		{
			name: "invalid config - empty host",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Port:     22,
					Username: "testuser",
					Method:   sftp.AuthPassword,
					Password: "testpass",
				},
			},
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - invalid port",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Host:     "example.com",
					Port:     -1,
					Username: "testuser",
					Method:   sftp.AuthPassword,
					Password: "testpass",
				},
			},
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - empty username",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Host:     "example.com",
					Port:     22,
					Method:   sftp.AuthPassword,
					Password: "testpass",
				},
			},
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
		{
			name: "invalid config - empty password for password auth",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Host:     "example.com",
					Port:     22,
					Username: "testuser",
					Method:   sftp.AuthPassword,
				},
			},
			expectError: true,
			errorType:   sftp.ErrAuthentication,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := sftp.NewClient(tt.config)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestNewClientWithDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name              string
		authHandler       sftp.AuthenticationHandler
		connectionManager sftp.ConnectionManager
		transferConfig    sftp.TransferConfig
		expectError       bool
		errorType         error
	}{
		{
			name:              "valid dependencies",
			authHandler:       sftp_mocks.NewMockAuthenticationHandler(ctrl),
			connectionManager: sftp_mocks.NewMockConnectionManager(ctrl),
			transferConfig:    sftp.DefaultConfig().Transfer,
			expectError:       false,
		},
		{
			name:              "nil auth handler",
			authHandler:       nil,
			connectionManager: sftp_mocks.NewMockConnectionManager(ctrl),
			transferConfig:    sftp.DefaultConfig().Transfer,
			expectError:       true,
			errorType:         sftp.ErrConfiguration,
		},
		{
			name:              "nil connection manager",
			authHandler:       sftp_mocks.NewMockAuthenticationHandler(ctrl),
			connectionManager: nil,
			transferConfig:    sftp.DefaultConfig().Transfer,
			expectError:       true,
			errorType:         sftp.ErrConfiguration,
		},
		{
			name:              "invalid transfer config - zero buffer size",
			authHandler:       sftp_mocks.NewMockAuthenticationHandler(ctrl),
			connectionManager: sftp_mocks.NewMockConnectionManager(ctrl),
			transferConfig: sftp.TransferConfig{
				BufferSize: 0,
			},
			expectError: false, // Zero value is ignored by merge, uses default
		},
		{
			name:              "invalid transfer config - buffer size too large",
			authHandler:       sftp_mocks.NewMockAuthenticationHandler(ctrl),
			connectionManager: sftp_mocks.NewMockConnectionManager(ctrl),
			transferConfig: sftp.TransferConfig{
				BufferSize: 20 * 1024 * 1024, // 20MB, exceeds 10MB limit
			},
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := sftp.NewClientWithDependencies(tt.authHandler, tt.connectionManager, tt.transferConfig)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClientConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
	mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)
	mockSFTPClient := &pkg_sftp.Client{}

	client, err := sftp.NewClientWithDependencies(mockAuthHandler, mockConnManager, sftp.DefaultConfig().Transfer)
	require.NoError(t, err)

	t.Run("successful connection", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		mockConnManager.EXPECT().
			GetConnection(ctx).
			Return(mockSFTPClient, nil).
			Times(1)

		mockConnManager.EXPECT().
			ReleaseConnection(mockSFTPClient).
			Return(nil).
			Times(1)

		err := client.Connect(ctx)
		require.NoError(t, err)

		// Second connect should be idempotent
		err = client.Connect(ctx)
		require.NoError(t, err)
	})

	t.Run("connection failure", func(t *testing.T) {
		mockAuthHandler2 := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager2 := sftp_mocks.NewMockConnectionManager(ctrl)

		client2, err := sftp.NewClientWithDependencies(mockAuthHandler2, mockConnManager2, sftp.DefaultConfig().Transfer)
		require.NoError(t, err)

		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		expectedErr := fmt.Errorf("%w: connection failed", sftp.ErrConnection)

		mockConnManager2.EXPECT().
			GetConnection(ctx).
			Return(nil, expectedErr).
			Times(1)

		err = client2.Connect(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrConnection)
	})
}

func TestClientClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("close before connect", func(t *testing.T) {
		mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)

		client, err := sftp.NewClientWithDependencies(mockAuthHandler, mockConnManager, sftp.DefaultConfig().Transfer)
		require.NoError(t, err)

		// Should not call Close on connection manager since not connected
		err = client.Close()
		require.NoError(t, err)
	})

	t.Run("close after connect", func(t *testing.T) {
		mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)
		mockSFTPClient := &pkg_sftp.Client{}

		client, err := sftp.NewClientWithDependencies(mockAuthHandler, mockConnManager, sftp.DefaultConfig().Transfer)
		require.NoError(t, err)

		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Connect first
		mockConnManager.EXPECT().
			GetConnection(ctx).
			Return(mockSFTPClient, nil)

		mockConnManager.EXPECT().
			ReleaseConnection(mockSFTPClient).
			Return(nil)

		err = client.Connect(ctx)
		require.NoError(t, err)

		// Now close
		mockConnManager.EXPECT().
			Close().
			Return(nil).
			Times(1)

		err = client.Close()
		require.NoError(t, err)
	})
}

func TestClientOperations_Integration(t *testing.T) {
	server := newTestSFTPServer(t)
	defer server.close()

	// Create client with real server
	config := sftp.Config{
		Authentication: sftp.AuthConfig{
			Host:            server.getAddress(),
			Port:            server.getPort(),
			Username:        server.auth.username,
			Method:          sftp.AuthPassword,
			Password:        server.auth.password,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	client, err := sftp.NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	err = client.Connect(context.Background())
	require.NoError(t, err)

	t.Run("Upload and Download", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a temporary local file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "upload-test.txt")
		content := []byte("test content for upload")
		err := os.WriteFile(localFile, content, 0644)
		require.NoError(t, err)

		// Upload the file
		remotePath := "upload-test.txt"
		err = client.Upload(ctx, localFile, remotePath)
		require.NoError(t, err)

		// Download the file
		downloadPath := filepath.Join(localDir, "download-test.txt")
		err = client.Download(ctx, remotePath, downloadPath)
		require.NoError(t, err)

		// Verify content
		downloadedContent, err := os.ReadFile(downloadPath)
		require.NoError(t, err)
		assert.Equal(t, content, downloadedContent)
	})

	t.Run("Upload with CreateDirs", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a temporary local file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "test.txt")
		err := os.WriteFile(localFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Upload to nested directory
		remotePath := "nested/dir/test.txt"
		err = client.Upload(ctx, localFile, remotePath, sftp.WithCreateDirs(true))
		require.NoError(t, err)

		// Verify file exists
		info, err := client.Stat(ctx, remotePath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
	})

	t.Run("Upload with OverwriteNever", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a temporary local file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "overwrite-test.txt")
		err := os.WriteFile(localFile, []byte("original"), 0644)
		require.NoError(t, err)

		remotePath := "overwrite-test.txt"

		// First upload should succeed
		err = client.Upload(ctx, localFile, remotePath)
		require.NoError(t, err)

		// Second upload with OverwriteNever should fail
		err = client.Upload(ctx, localFile, remotePath, sftp.WithUploadOverwriteNever())
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrDataTransfer)
	})

	t.Run("Download non-existent file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		localDir := t.TempDir()
		localPath := filepath.Join(localDir, "nonexistent.txt")

		err := client.Download(ctx, "nonexistent.txt", localPath)
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrFileNotFound)
	})

	t.Run("List directory", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create some test files
		err := client.Mkdir(ctx, "listtest")
		require.NoError(t, err)

		// Upload test files
		localDir := t.TempDir()
		for i := 1; i <= 3; i++ {
			localFile := filepath.Join(localDir, fmt.Sprintf("file%d.txt", i))
			err := os.WriteFile(localFile, []byte(fmt.Sprintf("content%d", i)), 0644)
			require.NoError(t, err)

			remotePath := fmt.Sprintf("listtest/file%d.txt", i)
			err = client.Upload(ctx, localFile, remotePath)
			require.NoError(t, err)
		}

		// List directory
		files, err := client.List(ctx, "listtest")
		require.NoError(t, err)
		assert.Len(t, files, 3)
	})

	t.Run("List non-existent directory", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		_, err := client.List(ctx, "nonexistent/dir")
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrDataTransfer)
	})

	t.Run("Mkdir", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		err := client.Mkdir(ctx, "testdir")
		require.NoError(t, err)

		// Verify directory exists
		info, err := client.Stat(ctx, "testdir")
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("Mkdir nested", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		err := client.Mkdir(ctx, "nested/test/dir")
		require.NoError(t, err)

		// Verify directory exists
		info, err := client.Stat(ctx, "nested/test/dir")
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("Remove file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "remove-test.txt")
		err := os.WriteFile(localFile, []byte("test"), 0644)
		require.NoError(t, err)

		remotePath := "remove-test.txt"
		err = client.Upload(ctx, localFile, remotePath)
		require.NoError(t, err)

		// Remove it
		err = client.Remove(ctx, remotePath)
		require.NoError(t, err)

		// Verify it's gone
		_, err = client.Stat(ctx, remotePath)
		require.Error(t, err)
	})

	t.Run("Rename file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "rename-test.txt")
		content := []byte("rename test")
		err := os.WriteFile(localFile, content, 0644)
		require.NoError(t, err)

		oldPath := "rename-old.txt"
		newPath := "rename-new.txt"

		err = client.Upload(ctx, localFile, oldPath)
		require.NoError(t, err)

		// Rename it
		err = client.Rename(ctx, oldPath, newPath)
		require.NoError(t, err)

		// Verify old path is gone
		_, err = client.Stat(ctx, oldPath)
		require.Error(t, err)

		// Verify new path exists
		info, err := client.Stat(ctx, newPath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
	})

	t.Run("Rename non-existent file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		err := client.Rename(ctx, "nonexistent.txt", "new.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrFileNotFound)
	})

	t.Run("Stat file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "stat-test.txt")
		content := []byte("stat test content")
		err := os.WriteFile(localFile, content, 0644)
		require.NoError(t, err)

		remotePath := "stat-test.txt"
		err = client.Upload(ctx, localFile, remotePath)
		require.NoError(t, err)

		// Stat it
		info, err := client.Stat(ctx, remotePath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
		assert.Equal(t, int64(len(content)), info.Size())
	})

	t.Run("Stat non-existent file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		_, err := client.Stat(ctx, "nonexistent.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrFileNotFound)
	})
}

func TestClientUploadOptions(t *testing.T) {
	server := newTestSFTPServer(t)
	defer server.close()

	config := sftp.Config{
		Authentication: sftp.AuthConfig{
			Host:            server.getAddress(),
			Port:            server.getPort(),
			Username:        server.auth.username,
			Method:          sftp.AuthPassword,
			Password:        server.auth.password,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	client, err := sftp.NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	err = client.Connect(context.Background())
	require.NoError(t, err)

	t.Run("Upload with progress callback", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a larger file to test progress
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "progress-test.txt")
		content := make([]byte, 1024*100) // 100KB
		for i := range content {
			content[i] = byte(i % 256)
		}
		err := os.WriteFile(localFile, content, 0644)
		require.NoError(t, err)

		var progressCalls int
		progressCallback := func(info sftp.ProgressInfo) {
			progressCalls++
			assert.GreaterOrEqual(t, info.BytesTransferred, int64(0))
			assert.Equal(t, int64(len(content)), info.TotalBytes)
			assert.GreaterOrEqual(t, info.Percentage, 0.0)
			assert.LessOrEqual(t, info.Percentage, 100.0)
		}

		remotePath := "progress-test.txt"
		err = client.Upload(ctx, localFile, remotePath, sftp.WithUploadProgress(progressCallback))
		require.NoError(t, err)
		assert.Greater(t, progressCalls, 0, "Progress callback should be called at least once")
	})

	t.Run("Upload with preserve permissions", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "perm-test.txt")
		err := os.WriteFile(localFile, []byte("test"), 0755)
		require.NoError(t, err)

		remotePath := "perm-test.txt"
		err = client.Upload(ctx, localFile, remotePath, sftp.WithPreservePermissions(true))
		require.NoError(t, err)

		// Verify file exists (permissions might not be preserved on all systems)
		info, err := client.Stat(ctx, remotePath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
	})
}

func TestClientDownloadOptions(t *testing.T) {
	server := newTestSFTPServer(t)
	defer server.close()

	config := sftp.Config{
		Authentication: sftp.AuthConfig{
			Host:            server.getAddress(),
			Port:            server.getPort(),
			Username:        server.auth.username,
			Method:          sftp.AuthPassword,
			Password:        server.auth.password,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	client, err := sftp.NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	err = client.Connect(context.Background())
	require.NoError(t, err)

	t.Run("Download with create dirs", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Upload a file first
		localDir := t.TempDir()
		uploadFile := filepath.Join(localDir, "upload.txt")
		content := []byte("test content")
		err := os.WriteFile(uploadFile, content, 0644)
		require.NoError(t, err)

		remotePath := "download-dirs-test.txt"
		err = client.Upload(ctx, uploadFile, remotePath)
		require.NoError(t, err)

		// Download to nested directory
		downloadPath := filepath.Join(localDir, "nested", "dir", "download.txt")
		err = client.Download(ctx, remotePath, downloadPath, sftp.WithDownloadCreateDirs(true))
		require.NoError(t, err)

		// Verify file exists
		downloadedContent, err := os.ReadFile(downloadPath)
		require.NoError(t, err)
		assert.Equal(t, content, downloadedContent)
	})

	t.Run("Download with progress callback", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Upload a larger file first
		localDir := t.TempDir()
		uploadFile := filepath.Join(localDir, "upload-progress.txt")
		content := make([]byte, 1024*50) // 50KB
		err := os.WriteFile(uploadFile, content, 0644)
		require.NoError(t, err)

		remotePath := "download-progress-test.txt"
		err = client.Upload(ctx, uploadFile, remotePath)
		require.NoError(t, err)

		var progressCalls int
		progressCallback := func(info sftp.ProgressInfo) {
			progressCalls++
		}

		downloadPath := filepath.Join(localDir, "download-progress.txt")
		err = client.Download(ctx, remotePath, downloadPath, sftp.WithDownloadProgress(progressCallback))
		require.NoError(t, err)
		assert.Greater(t, progressCalls, 0)
	})
}
