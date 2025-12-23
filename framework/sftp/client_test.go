package sftp_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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
			name: "valid private key authentication config with valid key data",
			config: sftp.Config{
				Authentication: sftp.AuthConfig{
					Host:            "example.com",
					Port:            22,
					Username:        "testuser",
					Method:          sftp.AuthPrivateKey,
					PrivateKeyData:  []byte("valid-key-data"),
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				},
			},
			expectError: false, // Config validation passes, key data validation happens at connect time
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
	tests := []struct {
		name              string
		authHandler       func(ctrl *gomock.Controller) sftp.AuthenticationHandler
		connectionManager func(ctrl *gomock.Controller) sftp.ConnectionManager
		transferConfig    sftp.TransferConfig
		expectError       bool
		errorType         error
	}{
		{
			name: "valid dependencies",
			authHandler: func(ctrl *gomock.Controller) sftp.AuthenticationHandler {
				return sftp_mocks.NewMockAuthenticationHandler(ctrl)
			},
			connectionManager: func(ctrl *gomock.Controller) sftp.ConnectionManager {
				return sftp_mocks.NewMockConnectionManager(ctrl)
			},
			transferConfig: sftp.DefaultConfig().Transfer,
			expectError:    false,
		},
		{
			name: "nil auth handler",
			authHandler: func(ctrl *gomock.Controller) sftp.AuthenticationHandler {
				return nil
			},
			connectionManager: func(ctrl *gomock.Controller) sftp.ConnectionManager {
				return sftp_mocks.NewMockConnectionManager(ctrl)
			},
			transferConfig: sftp.DefaultConfig().Transfer,
			expectError:    true,
			errorType:      sftp.ErrConfiguration,
		},
		{
			name: "nil connection manager",
			authHandler: func(ctrl *gomock.Controller) sftp.AuthenticationHandler {
				return sftp_mocks.NewMockAuthenticationHandler(ctrl)
			},
			connectionManager: func(ctrl *gomock.Controller) sftp.ConnectionManager {
				return nil
			},
			transferConfig: sftp.DefaultConfig().Transfer,
			expectError:    true,
			errorType:      sftp.ErrConfiguration,
		},
		{
			name: "invalid transfer config - zero buffer size",
			authHandler: func(ctrl *gomock.Controller) sftp.AuthenticationHandler {
				return sftp_mocks.NewMockAuthenticationHandler(ctrl)
			},
			connectionManager: func(ctrl *gomock.Controller) sftp.ConnectionManager {
				return sftp_mocks.NewMockConnectionManager(ctrl)
			},
			transferConfig: sftp.TransferConfig{
				BufferSize: 0,
			},
			expectError: false, // Zero value is ignored by merge, uses default
		},
		{
			name: "invalid transfer config - buffer size too large",
			authHandler: func(ctrl *gomock.Controller) sftp.AuthenticationHandler {
				return sftp_mocks.NewMockAuthenticationHandler(ctrl)
			},
			connectionManager: func(ctrl *gomock.Controller) sftp.ConnectionManager {
				return sftp_mocks.NewMockConnectionManager(ctrl)
			},
			transferConfig: sftp.TransferConfig{
				BufferSize: 20 * 1024 * 1024, // 20MB, exceeds 10MB limit
			},
			expectError: true,
			errorType:   sftp.ErrConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			authHandler := tt.authHandler(ctrl)
			connectionManager := tt.connectionManager(ctrl)
			client, err := sftp.NewClientWithDependencies(authHandler, connectionManager, tt.transferConfig)

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

func TestConnect(t *testing.T) {
	t.Run("should connect successfully", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)
		mockSFTPClient := &pkg_sftp.Client{}

		client, err := sftp.NewClientWithDependencies(
			mockAuthHandler,
			mockConnManager,
			sftp.DefaultConfig().Transfer,
		)
		require.NoError(t, err)

		mockConnManager.EXPECT().
			GetConnection(ctx).
			Return(mockSFTPClient, nil).
			Times(1)

		mockConnManager.EXPECT().
			ReleaseConnection(mockSFTPClient).
			Return(nil).
			Times(1)

		err = client.Connect(ctx)
		require.NoError(t, err)
	})

	t.Run("should return error when getting connection fails", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)

		client, err := sftp.NewClientWithDependencies(
			mockAuthHandler,
			mockConnManager,
			sftp.DefaultConfig().Transfer,
		)
		require.NoError(t, err)

		expectedErr := fmt.Errorf("%w: connection failed", sftp.ErrConnection)
		mockConnManager.EXPECT().
			GetConnection(ctx).
			Return(nil, expectedErr).
			Times(1)

		err = client.Connect(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrConnection)
	})

	t.Run("should return error when release connection fails", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)
		mockSFTPClient := &pkg_sftp.Client{}

		client, err := sftp.NewClientWithDependencies(
			mockAuthHandler,
			mockConnManager,
			sftp.DefaultConfig().Transfer,
		)
		require.NoError(t, err)

		mockConnManager.EXPECT().
			GetConnection(ctx).
			Return(mockSFTPClient, nil).
			Times(1)

		expectedErr := fmt.Errorf("%w: release failed", sftp.ErrConnection)
		mockConnManager.EXPECT().
			ReleaseConnection(mockSFTPClient).
			Return(expectedErr).
			Times(1)

		err = client.Connect(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrConnection)
	})
}

func TestClose(t *testing.T) {
	t.Run("should close connection successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAuthHandler := sftp_mocks.NewMockAuthenticationHandler(ctrl)
		mockConnManager := sftp_mocks.NewMockConnectionManager(ctrl)

		client, err := sftp.NewClientWithDependencies(mockAuthHandler, mockConnManager, sftp.DefaultConfig().Transfer)
		require.NoError(t, err)

		mockConnManager.EXPECT().
			Close().
			Return(nil).
			Times(1)

		err = client.Close()
		require.NoError(t, err)
	})
}

func TestUpload(t *testing.T) {
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

	t.Run("should upload file successfully", func(t *testing.T) {
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
	})

	t.Run("should return error for non-existent local file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Non-existent local file
		localFile := "nonexistent-file.txt"

		// Upload the file
		remotePath := "upload-test.txt"
		err := client.Upload(ctx, localFile, remotePath)
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrFileNotFound)
	})

	t.Run("should upload with CreateDirs", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a temporary local file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "test.txt")
		err := os.WriteFile(localFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Upload to nested directory
		remotePath := "nested/dir/test.txt"
		err = client.Upload(ctx, localFile, remotePath, sftp.WithUploadCreateDirs(true))
		require.NoError(t, err)
	})

	t.Run("should upload with ProgressCallback", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a temporary local file
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "test.txt")
		err := os.WriteFile(localFile, []byte("test"), 0644)
		require.NoError(t, err)

		progressCb := func(info sftp.ProgressInfo) {
			if info.Percentage == 100 {
				require.Equal(t, info.TotalBytes, info.BytesTransferred)
			}
		}
		remotePath := "progress-test.txt"
		err = client.Upload(ctx, localFile, remotePath, sftp.WithUploadProgress(progressCb))
		require.NoError(t, err)
	})

	t.Run("should upload with PreservePermissions", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Create a temporary local file with specific permissions
		localDir := t.TempDir()
		localFile := filepath.Join(localDir, "perm-test.txt")
		err := os.WriteFile(localFile, []byte("test"), 0755)
		require.NoError(t, err)

		// Upload the file
		remotePath := "perm-test.txt"
		err = client.Upload(ctx, localFile, remotePath, sftp.WithUploadPreservePermissions(true))
		require.NoError(t, err)

		// Verify file exists (permissions might not be preserved on all systems)
		info, err := client.Stat(ctx, remotePath)
		require.NoError(t, err)
		fileMode := info.Mode().Perm()
		assert.Equal(t, os.FileMode(0755), fileMode)
	})
}

func TestUpload_OverwritePolicy(t *testing.T) {
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

	// helper: create local file with content
	writeLocal := func(t *testing.T, dir, name string, content []byte) string {
		t.Helper()

		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, content, 0644))
		return p
	}

	// helper: download remote file content to verify overwrite happened
	readRemote := func(t *testing.T, remotePath string) []byte {
		t.Helper()
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		tmp := filepath.Join(t.TempDir(), "downloaded.txt")
		require.NoError(t, client.Download(ctx, remotePath, tmp))
		b, err := os.ReadFile(tmp)
		require.NoError(t, err)
		return b
	}

	// helper: set local file mod time
	setLocalModTime := func(t *testing.T, path string, mt time.Time) {
		t.Helper()
		// set both atime & mtime to mt
		require.NoError(t, os.Chtimes(path, mt, mt))
	}

	// helper: get remote file mod time
	getRemoteModTime := func(t *testing.T, remotePath string) time.Time {
		t.Helper()
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		info, err := client.Stat(ctx, remotePath)
		require.NoError(t, err)
		return info.ModTime()
	}

	t.Run("should return error when OverwriteNever and remote exists", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-never.txt"

		// First upload creates remote file
		local1 := writeLocal(t, localDir, "a.txt", []byte("first"))
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		// Second upload with OverwriteNever should fail
		local2 := writeLocal(t, localDir, "b.txt", []byte("second"))
		err := client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteNever())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		// Ensure remote content still equals first
		got := readRemote(t, remotePath)
		require.Equal(t, []byte("first"), got)
	})

	t.Run("should allow upload when remote missing even with OverwriteNever", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "not-exist-yet.txt"

		local := writeLocal(t, localDir, "a.txt", []byte("hello"))

		// Remote doesn't exist, so overwrite policy should not block
		err := client.Upload(ctx, local, remotePath, sftp.WithUploadOverwriteNever())
		require.NoError(t, err)

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("hello"), got)
	})

	t.Run("should upload and replace remote when OverwriteAlways", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-always.txt"

		local1 := writeLocal(t, localDir, "a.txt", []byte("first"))
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		local2 := writeLocal(t, localDir, "b.txt", []byte("second"))
		require.NoError(t, client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteAlways()))

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("second"), got)
	})

	t.Run("should return error when OverwriteIfDifferentSize and same size", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-diffsize.txt"

		// "AAAAA" length = 5
		local1 := writeLocal(t, localDir, "a.txt", []byte("AAAAA"))
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		// "BBBBB" length = 5 (same size) -> should fail
		local2 := writeLocal(t, localDir, "b.txt", []byte("BBBBB"))
		err := client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteIfDifferentSize())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("AAAAA"), got)
	})

	t.Run("should succeed when OverwriteIfDifferentSize and size differs", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-diffsize-ok.txt"

		local1 := writeLocal(t, localDir, "a.txt", []byte("AAAAA")) // 5
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		local2 := writeLocal(t, localDir, "b.txt", []byte("BBBBBB")) // 6
		require.NoError(t, client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteIfDifferentSize()))

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("BBBBBB"), got)
	})

	t.Run("should return error when OverwriteIfNewer and local is not newer", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-if-newer.txt"

		// First upload creates remote
		local1 := writeLocal(t, localDir, "a.txt", []byte("first"))
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		remoteMT := getRemoteModTime(t, remotePath)

		// Make local2 older than (or equal to) remote modtime
		local2 := writeLocal(t, localDir, "b.txt", []byte("second"))
		setLocalModTime(t, local2, remoteMT.Add(-2*time.Minute))

		err := client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteIfNewer())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		// Should remain first
		got := readRemote(t, remotePath)
		require.Equal(t, []byte("first"), got)
	})

	t.Run("should succeed when OverwriteIfNewer and local is newer", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-if-newer-ok.txt"

		local1 := writeLocal(t, localDir, "a.txt", []byte("first"))
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		remoteMT := getRemoteModTime(t, remotePath)

		local2 := writeLocal(t, localDir, "b.txt", []byte("second"))
		setLocalModTime(t, local2, remoteMT.Add(+2*time.Minute)) // newer than remote

		err := client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteIfNewer())
		require.NoError(t, err)

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("second"), got)
	})

	t.Run("should return error when OverwriteIfNewerOrDifferentSize and not newer and same size", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-newer-or-size.txt"

		// Create remote with size 5
		local1 := writeLocal(t, localDir, "a.txt", []byte("AAAAA"))
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		remoteMT := getRemoteModTime(t, remotePath)

		// Local2 same size (5) and older => should fail
		local2 := writeLocal(t, localDir, "b.txt", []byte("BBBBB"))
		setLocalModTime(t, local2, remoteMT.Add(-2*time.Minute))

		err := client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteIfNewerOrDifferentSize())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("AAAAA"), got)
	})

	t.Run("should succeed when OverwriteIfNewerOrDifferentSize and different size even if not newer", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		localDir := t.TempDir()
		remotePath := "overwrite-newer-or-size-size.txt"

		local1 := writeLocal(t, localDir, "a.txt", []byte("AAAAA")) // size 5
		require.NoError(t, client.Upload(ctx, local1, remotePath))

		remoteMT := getRemoteModTime(t, remotePath)

		local2 := writeLocal(t, localDir, "b.txt", []byte("BBBBBB")) // size 6 different
		setLocalModTime(t, local2, remoteMT.Add(-2*time.Minute))     // older, but size differs

		err := client.Upload(ctx, local2, remotePath, sftp.WithUploadOverwriteIfNewerOrDifferentSize())
		require.NoError(t, err)

		got := readRemote(t, remotePath)
		require.Equal(t, []byte("BBBBBB"), got)
	})
}

func TestDownload(t *testing.T) {
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

	// helper: create local file
	writeLocal := func(t *testing.T, dir, name string, content []byte, perm os.FileMode) string {
		t.Helper()
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, content, perm))
		return p
	}

	// helper: read local
	readLocal := func(t *testing.T, path string) []byte {
		t.Helper()
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		return b
	}

	t.Run("should download file successfully", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("hello"), 0644)
		remotePath := "download-ok.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Download to dst
		dst := filepath.Join(t.TempDir(), "out.txt")
		require.NoError(t, client.Download(ctx, remotePath, dst))
		require.Equal(t, []byte("hello"), readLocal(t, dst))
	})

	t.Run("should return error when remote file not found", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		dst := filepath.Join(t.TempDir(), "out.txt")
		err := client.Download(ctx, "no-such-file.txt", dst)
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrFileNotFound)
	})

	t.Run("should download with CreateDirs", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("nested"), 0644)
		remotePath := "download-createdirs.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Download to nested dirs
		dst := filepath.Join(t.TempDir(), "a/b/c/out.txt")
		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadCreateDirs(true)))
		require.Equal(t, []byte("nested"), readLocal(t, dst))
	})

	t.Run("should download with ProgressCallback", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("progress"), 0644)
		remotePath := "download-progress.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Download to dst
		progressCb := func(info sftp.ProgressInfo) {
			if info.Percentage == 100 {
				require.Equal(t, info.TotalBytes, info.BytesTransferred)
			}
		}
		dst := filepath.Join(t.TempDir(), "out.txt")
		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadProgress(progressCb)))
		require.Equal(t, []byte("progress"), readLocal(t, dst))
	})

	t.Run("should download with PreservePermissions", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file with specific permissions
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("perms"), 0755)
		remotePath := "download-perms.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath, sftp.WithUploadPreservePermissions(true)))

		// Download to dst
		dst := filepath.Join(t.TempDir(), "out.txt")
		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadPreservePermissions(true)))

		info, err := os.Stat(dst)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0755), info.Mode().Perm())
	})
}

func TestDownload_OverwritePolicy(t *testing.T) {
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

	// helper: create local file
	writeLocal := func(t *testing.T, dir, name string, content []byte, perm os.FileMode) string {
		t.Helper()
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, content, perm))
		return p
	}

	// helper: read local
	readLocal := func(t *testing.T, path string) []byte {
		t.Helper()
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		return b
	}

	// helper: set local mod time (for overwrite-if-newer tests)
	setLocalModTime := func(t *testing.T, path string, mt time.Time) {
		t.Helper()
		require.NoError(t, os.Chtimes(path, mt, mt))
	}

	t.Run("should return error when OverwriteNever and local exists", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("REMOTE"), 0644)
		remotePath := "download-overwrite-never.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Create local already exists
		dst := writeLocal(t, t.TempDir(), "out.txt", []byte("LOCAL"), 0644)

		err := client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteNever())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		// Still local content
		require.Equal(t, []byte("LOCAL"), readLocal(t, dst))
	})

	t.Run("should replace local when OverwriteAlways", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("REMOTE"), 0644)
		remotePath := "download-overwrite-always.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Create local already exists
		dst := writeLocal(t, t.TempDir(), "out.txt", []byte("LOCAL"), 0644)

		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteAlways()))
		require.Equal(t, []byte("REMOTE"), readLocal(t, dst))
	})

	t.Run("should return error when OverwriteIfDifferentSize and same size", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Remote size 5
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("AAAAA"), 0644)
		remotePath := "download-overwrite-diffsize.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Local size 5
		dst := writeLocal(t, t.TempDir(), "out.txt", []byte("BBBBB"), 0644)

		err := client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfDifferentSize())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		require.Equal(t, []byte("BBBBB"), readLocal(t, dst))
	})

	t.Run("should replace local when OverwriteIfDifferentSize and size differs", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Remote size 6
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("AAAAAA"), 0644)
		remotePath := "download-overwrite-diffsize-ok.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Local size 5
		dst := writeLocal(t, t.TempDir(), "out.txt", []byte("BBBBB"), 0644)

		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfDifferentSize()))
		require.Equal(t, []byte("AAAAAA"), readLocal(t, dst))
	})

	t.Run("should return error when OverwriteIfNewer and remote is not newer (remote older/equal)", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("REMOTE"), 0644)
		remotePath := "download-overwrite-if-newer.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Create local and make it FUTURE so remote is not newer
		dst := writeLocal(t, t.TempDir(), "out.txt", []byte("LOCAL"), 0644)
		setLocalModTime(t, dst, time.Now().Add(+5*time.Minute))

		err := client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfNewer())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		require.Equal(t, []byte("LOCAL"), readLocal(t, dst))
	})

	t.Run("should replace local when OverwriteIfNewer and remote is newer than local", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		// Seed remote file
		src := writeLocal(t, t.TempDir(), "seed.txt", []byte("REMOTE"), 0644)
		remotePath := "download-overwrite-if-newer-ok.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Create local and make it PAST so remote is newer
		dst := writeLocal(t, t.TempDir(), "out.txt", []byte("LOCAL"), 0644)
		setLocalModTime(t, dst, time.Now().Add(-5*time.Minute))

		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfNewer()))
		require.Equal(t, []byte("REMOTE"), readLocal(t, dst))
	})

	t.Run("should return error when OverwriteIfNewerOrDifferentSize and remote not newer and same size", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		localDir := t.TempDir()

		// Remote size 5
		src := writeLocal(t, localDir, "seed.txt", []byte("AAAAA"), 0644)
		remotePath := "download-overwrite-newer-or-size.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Local size 5 and FUTURE (remote not newer)
		dstDir := t.TempDir()
		dst := writeLocal(t, dstDir, "out.txt", []byte("BBBBB"), 0644)
		setLocalModTime(t, dst, time.Now().Add(+5*time.Minute))

		err := client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfNewerOrDifferentSize())
		require.Error(t, err)
		require.ErrorIs(t, err, sftp.ErrDataTransfer)

		require.Equal(t, []byte("BBBBB"), readLocal(t, dst))
	})

	t.Run("should replace local when OverwriteIfNewerOrDifferentSize and size differs even if remote not newer", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		localDir := t.TempDir()

		// Remote size 6
		src := writeLocal(t, localDir, "seed.txt", []byte("AAAAAA"), 0644)
		remotePath := "download-overwrite-newer-or-size-size.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Local size 5 and FUTURE (remote not newer) -> still allowed because size differs
		dstDir := t.TempDir()
		dst := writeLocal(t, dstDir, "out.txt", []byte("BBBBB"), 0644)
		setLocalModTime(t, dst, time.Now().Add(+5*time.Minute))

		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfNewerOrDifferentSize()))
		require.Equal(t, []byte("AAAAAA"), readLocal(t, dst))
	})

	t.Run("should replace local when OverwriteIfNewerOrDifferentSize and remote is newer even if same size", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())
		localDir := t.TempDir()

		// Remote size 5
		src := writeLocal(t, localDir, "seed.txt", []byte("AAAAA"), 0644)
		remotePath := "download-overwrite-newer-or-size-newer.txt"
		require.NoError(t, client.Upload(ctx, src, remotePath))

		// Local size 5 but older -> remote newer => allowed
		dstDir := t.TempDir()
		dst := writeLocal(t, dstDir, "out.txt", []byte("BBBBB"), 0644)
		setLocalModTime(t, dst, time.Now().Add(-5*time.Minute))

		require.NoError(t, client.Download(ctx, remotePath, dst, sftp.WithDownloadOverwriteIfNewerOrDifferentSize()))
		require.Equal(t, []byte("AAAAA"), readLocal(t, dst))
	})
}

func TestList(t *testing.T) {
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

	t.Run("should list directory successfully", func(t *testing.T) {
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

	t.Run("should return error for non-existent directory", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		_, err := client.List(ctx, "nonexistent/dir")
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrDataTransfer)
	})
}

func TestMkdir(t *testing.T) {
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

	t.Run("should create directory successfully", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		err := client.Mkdir(ctx, "testdir")
		require.NoError(t, err)

		// Verify directory exists
		info, err := client.Stat(ctx, "testdir")
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("should create nested directory successfully", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		err := client.Mkdir(ctx, "nested/test/dir")
		require.NoError(t, err)

		// Verify directory exists
		info, err := client.Stat(ctx, "nested/test/dir")
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestRemove(t *testing.T) {
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

	t.Run("should remove file successfully", func(t *testing.T) {
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
}

func TestRename(t *testing.T) {
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

	t.Run("should rename file successfully", func(t *testing.T) {
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

	t.Run("should return error when renaming non-existent file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		err := client.Rename(ctx, "nonexistent.txt", "new.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrFileNotFound)
	})
}

func TestStat(t *testing.T) {
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

	t.Run("should stat file successfully", func(t *testing.T) {
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

	t.Run("should return error when statting non-existent file", func(t *testing.T) {
		ctx := logger.NewContext(context.Background(), logger.NewNoopLogger())

		_, err := client.Stat(ctx, "nonexistent.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, sftp.ErrFileNotFound)
	})
}
