package sftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	common_logger "github.com/kittipat1413/go-common/framework/logger"
	"github.com/pkg/sftp"
)

//go:generate mockgen -source=./client.go -destination=./mocks/client.go -package=sftp_mocks

// Client is the main interface for SFTP operations
type Client interface {
	// Connect establishes a new SFTP connection and registers it in the connection pool
	Connect(ctx context.Context) error
	// Upload uploads a local file to the remote SFTP server
	Upload(ctx context.Context, localPath, remotePath string, opts ...UploadOption) error
	// Download downloads a remote file from the SFTP server to local storage
	Download(ctx context.Context, remotePath, localPath string, opts ...DownloadOption) error
	// List lists files and directories in the specified remote path
	List(ctx context.Context, remotePath string) ([]os.FileInfo, error)
	// Mkdir creates a directory and all necessary parent directories on the remote SFTP server
	Mkdir(ctx context.Context, remotePath string) error
	// Remove removes a file or directory and all its contents from the remote SFTP server
	Remove(ctx context.Context, remotePath string) error
	// Rename renames or moves a file or directory on the remote SFTP server
	Rename(ctx context.Context, oldPath, newPath string) error
	// Stat returns information about a file or directory on the remote SFTP server
	Stat(ctx context.Context, remotePath string) (os.FileInfo, error)
	// Close forcefully closes all connections in the SFTP connection pool, including connections currently in use.
	// This method is intended to be called during application shutdown
	Close() error
}

// sftpClient is the concrete implementation of the Client interface
type sftpClient struct {
	authHandler       AuthenticationHandler
	connectionManager ConnectionManager
	transferConfig    TransferConfig
	connected         bool
}

// NewClient creates a new SFTP client with the given configuration
func NewClient(config Config) (Client, error) {
	// Merge with defaults and validate
	mergedConfig := MergeConfig(config)
	if err := validateConfig(mergedConfig); err != nil {
		return nil, err
	}

	// Create authentication handler
	authHandler, err := CreateAuthHandler(mergedConfig.Authentication)
	if err != nil {
		return nil, err // errors are wrapped in CreateAuthHandler
	}

	// Create connection manager
	connectionManager, err := NewConnectionManager(authHandler, mergedConfig.Authentication, mergedConfig.Connection)
	if err != nil {
		return nil, err // errors are wrapped in NewConnectionManager
	}

	return NewClientWithDependencies(authHandler, connectionManager, mergedConfig.Transfer)
}

// NewClientWithDependencies creates a new SFTP client with injected dependencies
// This constructor is useful for testing or when you need more control over the
// ConnectionManager and AuthenticationHandler implementations
func NewClientWithDependencies(authHandler AuthenticationHandler, connectionManager ConnectionManager, transferConfig TransferConfig) (Client, error) {
	// Validate dependencies
	if connectionManager == nil {
		return nil, fmt.Errorf("%w: connection manager cannot be nil", ErrConfiguration)
	}
	if authHandler == nil {
		return nil, fmt.Errorf("%w: authentication handler cannot be nil", ErrConfiguration)
	}

	// Merge with defaults and validate transfer config
	mergedTransferConfig := mergeTransferConfig(DefaultConfig().Transfer, transferConfig)
	if err := validateTransferConfig(mergedTransferConfig); err != nil {
		return nil, err
	}

	return &sftpClient{
		authHandler:       authHandler,
		connectionManager: connectionManager,
		transferConfig:    mergedTransferConfig,
		connected:         false,
	}, nil
}

// Connect establishes a new SFTP connection and registers it in the connection pool
func (c *sftpClient) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Test connection by getting a client from the pool
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return err // errors are wrapped in GetConnection
	}

	// Release the connection back to the pool
	if err := c.connectionManager.ReleaseConnection(client); err != nil {
		return err // errors are wrapped in ReleaseConnection
	}

	c.connected = true
	return nil
}

// Close forcefully closes all connections in the SFTP connection pool, including connections currently in use
// This method is intended to be called during application shutdown
func (c *sftpClient) Close() error {
	if !c.connected {
		return nil
	}

	if err := c.connectionManager.Close(); err != nil {
		return err // errors are wrapped in Close
	}

	c.connected = false
	return nil
}

// UploadConfig configures how Upload behaves
type UploadConfig struct {
	// CreateDirs controls whether Upload should create remote parent directories
	CreateDirs bool
	// PreservePermissions controls whether Upload should attempt to copy local file mode
	// to the remote file after transfer
	PreservePermissions bool
	// ProgressCallback, if set, is invoked as bytes are transferred
	ProgressCallback ProgressCallback
	// OverwritePolicy controls how Upload behaves if the remote path already exists
	OverwritePolicy OverwritePolicy
}

// UploadOption defines options for upload operations
type UploadOption func(*UploadConfig)

// WithCreateDirs sets whether to create directories during upload
func WithCreateDirs(create bool) UploadOption {
	return func(config *UploadConfig) {
		config.CreateDirs = create
	}
}

// WithPreservePermissions sets whether to preserve file permissions during upload
func WithPreservePermissions(preserve bool) UploadOption {
	return func(config *UploadConfig) {
		config.PreservePermissions = preserve
	}
}

// WithUploadProgress sets a progress callback for upload operations
func WithUploadProgress(callback ProgressCallback) UploadOption {
	return func(config *UploadConfig) {
		config.ProgressCallback = callback
	}
}

// WithUploadOverwritePolicy sets the overwrite policy for upload operations
func WithUploadOverwritePolicy(policy OverwritePolicy) UploadOption {
	return func(config *UploadConfig) {
		config.OverwritePolicy = policy
	}
}

// Upload uploads a local file to the remote SFTP server
//
// Behavior:
//   - If CreateDirs is enabled, Upload creates the remote parent directories
//     before uploading
//   - Overwrite behavior is controlled by OverwritePolicy (default: OverwriteAlways),
//     which may skip, error, or replace depending on the policy
//   - If a ProgressCallback is provided, Upload reports incremental progress while
//     transferring the file
//   - If PreservePermissions is enabled, Upload attempts to apply the local file mode
//     to the remote file (failure is logged as a warning and does not fail the upload)
func (c *sftpClient) Upload(ctx context.Context, localPath, remotePath string, opts ...UploadOption) error {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Apply options
	config := &UploadConfig{
		CreateDirs:          c.transferConfig.CreateDirs,
		PreservePermissions: c.transferConfig.PreservePermissions,
		ProgressCallback:    c.transferConfig.ProgressCallback,
		OverwritePolicy:     OverwriteAlways,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("%w: failed to open local file %s: %v", ErrFileNotFound, localPath, err)
	}
	defer localFile.Close()

	// Get local file info
	localInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("%w: failed to stat local file %s: %v", ErrFileNotFound, localPath, err)
	}

	// Create remote directory if needed
	if config.CreateDirs {
		remoteDir := filepath.Dir(remotePath)
		if remoteDir != "." && remoteDir != "/" {
			if err := c.createRemoteDir(client, remoteDir); err != nil {
				return err
			}
		}
	}

	// Check overwrite policy
	if err := c.checkOverwritePolicy(client, remotePath, localInfo, config.OverwritePolicy); err != nil {
		return err
	}

	// Create remote file
	remoteFile, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("%w: failed to create remote file %s: %v", ErrDataTransfer, remotePath, err)
	}
	defer remoteFile.Close()

	// Copy file with progress tracking
	if err := c.copyWithProgress(localFile, remoteFile, localInfo.Size(), config.ProgressCallback); err != nil {
		return fmt.Errorf("%w: failed to transfer file to %s: %v", ErrDataTransfer, remotePath, err)
	}

	// Preserve permissions if requested
	if config.PreservePermissions {
		if err := client.Chmod(remotePath, localInfo.Mode()); err != nil {
			logger.Warn(ctx, "Failed to set file permissions", common_logger.Fields{
				"operation":   "upload",
				"remote_path": remotePath,
				"permissions": localInfo.Mode(),
				"error":       err.Error(),
			})
		}
	}

	logger.Debug(ctx, "File upload completed successfully", common_logger.Fields{
		"operation":     "upload",
		"local_path":    localPath,
		"remote_path":   remotePath,
		"file_size":     localInfo.Size(),
		"duration":      time.Since(startTime),
		"transfer_rate": fmt.Sprintf("%.2f MB/s", float64(localInfo.Size())/(1024*1024)/time.Since(startTime).Seconds()),
	})

	return nil
}

// DownloadConfig configures how Download behaves
type DownloadConfig struct {
	// CreateDirs controls whether Download should create local parent directories
	CreateDirs bool
	// PreservePermissions controls whether Download should attempt to copy the remote
	// file mode to the local file after transfer
	PreservePermissions bool
	// ProgressCallback, if set, is invoked as bytes are transferred
	ProgressCallback ProgressCallback
	// OverwritePolicy controls how Download behaves if the local path already exists
	OverwritePolicy OverwritePolicy
}

// DownloadOption defines options for download operations
type DownloadOption func(*DownloadConfig)

// WithDownloadCreateDirs sets whether to create directories during download
func WithDownloadCreateDirs(create bool) DownloadOption {
	return func(config *DownloadConfig) {
		config.CreateDirs = create
	}
}

// WithDownloadPreservePermissions sets whether to preserve file permissions during download
func WithDownloadPreservePermissions(preserve bool) DownloadOption {
	return func(config *DownloadConfig) {
		config.PreservePermissions = preserve
	}
}

// WithDownloadProgress sets a progress callback for download operations
func WithDownloadProgress(callback ProgressCallback) DownloadOption {
	return func(config *DownloadConfig) {
		config.ProgressCallback = callback
	}
}

// WithDownloadOverwritePolicy sets the overwrite policy for download operations
func WithDownloadOverwritePolicy(policy OverwritePolicy) DownloadOption {
	return func(config *DownloadConfig) {
		config.OverwritePolicy = policy
	}
}

// Download downloads a remote file from the SFTP server to local storage
//
// Behavior:
//   - If CreateDirs is enabled, Download creates the local parent directories
//     before writing the file
//   - Overwrite behavior is controlled by OverwritePolicy (default: OverwriteAlways),
//     which may skip, error, or replace depending on the policy
//   - If a ProgressCallback is provided, Download reports incremental progress while
//     transferring the file
//   - If PreservePermissions is enabled, Download attempts to apply the remote file mode
//     to the local file after transfer (failure is logged as a warning and does not fail
//     the download)
func (c *sftpClient) Download(ctx context.Context, remotePath, localPath string, opts ...DownloadOption) error {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Apply options
	config := &DownloadConfig{
		CreateDirs:          c.transferConfig.CreateDirs,
		PreservePermissions: c.transferConfig.PreservePermissions,
		ProgressCallback:    c.transferConfig.ProgressCallback,
		OverwritePolicy:     OverwriteAlways,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Open remote file
	remoteFile, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("%w: failed to open remote file %s: %v", ErrFileNotFound, remotePath, err)
	}
	defer remoteFile.Close()

	// Get remote file info
	remoteInfo, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("%w: failed to stat remote file %s: %v", ErrFileNotFound, remotePath, err)
	}

	// Create local directory if needed
	if config.CreateDirs {
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return fmt.Errorf("%w: failed to create local directory %s: %v", ErrDataTransfer, localDir, err)
		}
	}

	// Check overwrite policy
	if err := c.checkLocalOverwritePolicy(localPath, remoteInfo, config.OverwritePolicy); err != nil {
		return err
	}

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("%w: failed to create local file %s: %v", ErrDataTransfer, localPath, err)
	}
	defer localFile.Close()

	// Copy file with progress tracking
	if err := c.copyWithProgress(remoteFile, localFile, remoteInfo.Size(), config.ProgressCallback); err != nil {
		return fmt.Errorf("%w: failed to transfer file to %s: %v", ErrDataTransfer, localPath, err)
	}

	// Preserve permissions if requested
	if config.PreservePermissions {
		if err := os.Chmod(localPath, remoteInfo.Mode()); err != nil {
			logger.Warn(ctx, "Failed to set file permissions", common_logger.Fields{
				"operation":   "download",
				"local_path":  localPath,
				"permissions": remoteInfo.Mode(),
				"error":       err.Error(),
			})
		}
	}

	logger.Debug(ctx, "File download completed successfully", common_logger.Fields{
		"operation":     "download",
		"remote_path":   remotePath,
		"local_path":    localPath,
		"file_size":     remoteInfo.Size(),
		"duration":      time.Since(startTime),
		"transfer_rate": fmt.Sprintf("%.2f MB/s", float64(remoteInfo.Size())/(1024*1024)/time.Since(startTime).Seconds()),
	})

	return nil
}

// List lists files and directories in the specified remote path
func (c *sftpClient) List(ctx context.Context, remotePath string) ([]os.FileInfo, error) {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return nil, err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Read directory
	entries, err := client.ReadDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list directory %s: %v", ErrDataTransfer, remotePath, err)
	}

	logger.Debug(ctx, "Directory listing completed successfully", common_logger.Fields{
		"operation":   "list",
		"remote_path": remotePath,
		"entry_count": len(entries),
		"duration":    time.Since(startTime),
	})

	return entries, nil
}

// Mkdir creates a directory and all necessary parent directories on the remote SFTP server
func (c *sftpClient) Mkdir(ctx context.Context, remotePath string) error {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Create directory
	if err := client.MkdirAll(remotePath); err != nil {
		return fmt.Errorf("%w: failed to create directory %s: %v", ErrDataTransfer, remotePath, err)
	}

	logger.Debug(ctx, "Directory created successfully", common_logger.Fields{
		"operation":   "mkdir",
		"remote_path": remotePath,
		"duration":    time.Since(startTime),
	})

	return nil
}

// Remove removes a file or directory and all its contents from the remote SFTP server
func (c *sftpClient) Remove(ctx context.Context, remotePath string) error {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Remove file or directory
	err = client.RemoveAll(remotePath)
	if err != nil {
		return fmt.Errorf("%w: failed to remove file %s: %v", ErrDataTransfer, remotePath, err)
	}

	logger.Debug(ctx, "Path removed successfully", common_logger.Fields{
		"operation":   "remove",
		"remote_path": remotePath,
		"duration":    time.Since(startTime),
	})

	return nil
}

// Rename renames or moves a file or directory on the remote SFTP server
// This operation is atomic where supported by the server
func (c *sftpClient) Rename(ctx context.Context, oldPath, newPath string) error {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Verify source exists before attempting rename
	sourceInfo, err := client.Stat(oldPath)
	if err != nil {
		return fmt.Errorf("%w: source path does not exist %s: %v", ErrFileNotFound, oldPath, err)
	}

	// Create destination directory if needed
	newDir := filepath.Dir(newPath)
	if newDir != "." && newDir != "/" {
		if err := c.createRemoteDir(client, newDir); err != nil {
			return fmt.Errorf("%w: failed to create destination directory: %v", ErrDataTransfer, err)
		}
	}

	// Perform atomic rename/move operation
	if err := client.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("%w: failed to rename/move from %s to %s: %v", ErrDataTransfer, oldPath, newPath, err)
	}

	logger.Debug(ctx, "Path renamed successfully", common_logger.Fields{
		"operation": "rename",
		"old_path":  oldPath,
		"new_path":  newPath,
		"type":      map[bool]string{true: "directory", false: "file"}[sourceInfo.IsDir()],
		"duration":  time.Since(startTime),
	})

	return nil
}

// Stat returns information about a file or directory on the remote SFTP server
func (c *sftpClient) Stat(ctx context.Context, remotePath string) (os.FileInfo, error) {
	startTime := time.Now()
	logger := common_logger.FromContext(ctx)

	// Get connection
	client, err := c.connectionManager.GetConnection(ctx)
	if err != nil {
		return nil, err // errors are wrapped in GetConnection
	}
	defer func() {
		_ = c.connectionManager.ReleaseConnection(client)
	}()

	// Get file info
	info, err := client.Stat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to stat path %s: %v", ErrFileNotFound, remotePath, err)
	}

	logger.Debug(ctx, "File information retrieved successfully", common_logger.Fields{
		"operation":   "stat",
		"remote_path": remotePath,
		"type":        map[bool]string{true: "directory", false: "file"}[info.IsDir()],
		"size":        info.Size(),
		"mode":        info.Mode(),
		"duration":    time.Since(startTime),
	})

	return info, nil
}

// OverwritePolicy defines how to handle existing files
type OverwritePolicy int

const (
	OverwriteAlways OverwritePolicy = iota
	OverwriteNever
	OverwriteIfNewer
	OverwriteIfDifferentSize
	OverwriteIfNewerOrDifferentSize
)

// ProgressInfo contains information about transfer progress
type ProgressInfo struct {
	BytesTransferred int64
	TotalBytes       int64
	Percentage       float64
	Speed            int64 // bytes per second
}

// ProgressCallback is called during file transfers to report progress
type ProgressCallback func(info ProgressInfo)

// Helper methods

// copyWithProgress copies data from src to dst with optional progress tracking
func (c *sftpClient) copyWithProgress(src io.Reader, dst io.Writer, totalBytes int64, progressCallback ProgressCallback) error {
	buffer := make([]byte, c.transferConfig.BufferSize)

	var bytesTransferred int64
	startTime := time.Now()
	lastProgressTime := startTime

	// Call initial progress callback
	if progressCallback != nil {
		progressCallback(ProgressInfo{
			BytesTransferred: 0,
			TotalBytes:       totalBytes,
			Percentage:       0,
			Speed:            0,
		})
	}

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			bytesTransferred += int64(n)

			// Call progress callback if provided and enough time has passed
			// Throttle progress callbacks to avoid overwhelming the callback
			now := time.Now()
			if progressCallback != nil && now.Sub(lastProgressTime) >= 30*time.Second {
				elapsed := now.Sub(startTime)
				var speed int64
				if elapsed.Seconds() > 0 {
					speed = int64(float64(bytesTransferred) / elapsed.Seconds())
				}

				var percentage float64
				if totalBytes > 0 {
					percentage = float64(bytesTransferred) / float64(totalBytes) * 100
					// Ensure percentage doesn't exceed 100%
					if percentage > 100 {
						percentage = 100
					}
				}

				progressCallback(ProgressInfo{
					BytesTransferred: bytesTransferred,
					TotalBytes:       totalBytes,
					Percentage:       percentage,
					Speed:            speed,
				})
				lastProgressTime = now
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Final progress callback to ensure 100% completion is reported
	if progressCallback != nil && totalBytes > 0 {
		elapsed := time.Since(startTime)
		var speed int64
		if elapsed.Seconds() > 0 {
			speed = int64(float64(bytesTransferred) / elapsed.Seconds())
		}

		progressCallback(ProgressInfo{
			BytesTransferred: bytesTransferred,
			TotalBytes:       totalBytes,
			Percentage:       100,
			Speed:            speed,
		})
	}

	return nil
}

// createRemoteDir creates a remote directory recursively
func (c *sftpClient) createRemoteDir(client *sftp.Client, remotePath string) error {
	// Clean the path
	remotePath = filepath.Clean(remotePath)
	if remotePath == "." || remotePath == "/" {
		return nil
	}

	// Check if directory already exists
	if _, err := client.Stat(remotePath); err == nil {
		return nil // Directory already exists
	}

	// Create parent directory first
	parent := filepath.Dir(remotePath)
	if parent != "." && parent != "/" && parent != remotePath {
		if err := c.createRemoteDir(client, parent); err != nil {
			return err
		}
	}

	// Create this directory
	if err := client.Mkdir(remotePath); err != nil {
		return fmt.Errorf("%w: failed to create directory %s: %v", ErrDataTransfer, remotePath, err)
	}

	return nil
}

// checkOverwritePolicy checks if a remote file can be overwritten based on policy
func (c *sftpClient) checkOverwritePolicy(client *sftp.Client, remotePath string, localInfo os.FileInfo, policy OverwritePolicy) error {
	if policy == OverwriteAlways {
		return nil
	}

	// Check if remote file exists
	remoteInfo, err := client.Stat(remotePath)
	if err != nil {
		// File doesn't exist, so we can create it
		return nil
	}

	switch policy {
	case OverwriteNever:
		return fmt.Errorf("%w: file %s already exists and overwrite policy is never", ErrDataTransfer, remotePath)
	case OverwriteIfNewer:
		if localInfo.ModTime().Before(remoteInfo.ModTime()) || localInfo.ModTime().Equal(remoteInfo.ModTime()) {
			return fmt.Errorf("%w: local file is not newer than remote file %s", ErrDataTransfer, remotePath)
		}
	case OverwriteIfDifferentSize:
		if localInfo.Size() == remoteInfo.Size() {
			return fmt.Errorf("%w: local and remote files have the same size for %s", ErrDataTransfer, remotePath)
		}
	case OverwriteIfNewerOrDifferentSize:
		isNewer := localInfo.ModTime().After(remoteInfo.ModTime())
		isDifferentSize := localInfo.Size() != remoteInfo.Size()
		if !isNewer && !isDifferentSize {
			return fmt.Errorf("%w: local file is not newer and has the same size as remote file %s", ErrDataTransfer, remotePath)
		}
	}

	return nil
}

// checkLocalOverwritePolicy checks if a local file can be overwritten based on policy
func (c *sftpClient) checkLocalOverwritePolicy(localPath string, remoteInfo os.FileInfo, policy OverwritePolicy) error {
	if policy == OverwriteAlways {
		return nil
	}

	// Check if local file exists
	localInfo, err := os.Stat(localPath)
	if err != nil {
		// File doesn't exist, so we can create it
		return nil
	}

	switch policy {
	case OverwriteNever:
		return fmt.Errorf("%w: file %s already exists and overwrite policy is never", ErrDataTransfer, localPath)
	case OverwriteIfNewer:
		if remoteInfo.ModTime().Before(localInfo.ModTime()) || remoteInfo.ModTime().Equal(localInfo.ModTime()) {
			return fmt.Errorf("%w: remote file is not newer than local file %s", ErrDataTransfer, localPath)
		}
	case OverwriteIfDifferentSize:
		if remoteInfo.Size() == localInfo.Size() {
			return fmt.Errorf("%w: remote and local files have the same size for %s", ErrDataTransfer, localPath)
		}
	case OverwriteIfNewerOrDifferentSize:
		isNewer := remoteInfo.ModTime().After(localInfo.ModTime())
		isDifferentSize := remoteInfo.Size() != localInfo.Size()
		if !isNewer && !isDifferentSize {
			return fmt.Errorf("%w: remote file is not newer and has the same size as local file %s", ErrDataTransfer, localPath)
		}
	}

	return nil
}

// Convenience functions for common overwrite policies

// WithUploadOverwriteAlways sets upload to always overwrite existing files
func WithUploadOverwriteAlways() UploadOption {
	return WithUploadOverwritePolicy(OverwriteAlways)
}

// WithUploadOverwriteNever sets upload to never overwrite existing files
func WithUploadOverwriteNever() UploadOption {
	return WithUploadOverwritePolicy(OverwriteNever)
}

// WithUploadOverwriteIfNewer sets upload to overwrite only if local file is newer
func WithUploadOverwriteIfNewer() UploadOption {
	return WithUploadOverwritePolicy(OverwriteIfNewer)
}

// WithUploadOverwriteIfDifferentSize sets upload to overwrite only if file sizes differ
func WithUploadOverwriteIfDifferentSize() UploadOption {
	return WithUploadOverwritePolicy(OverwriteIfDifferentSize)
}

// WithUploadOverwriteIfNewerOrDifferentSize sets upload to overwrite if newer or different size
func WithUploadOverwriteIfNewerOrDifferentSize() UploadOption {
	return WithUploadOverwritePolicy(OverwriteIfNewerOrDifferentSize)
}

// WithDownloadOverwriteAlways sets download to always overwrite existing files
func WithDownloadOverwriteAlways() DownloadOption {
	return WithDownloadOverwritePolicy(OverwriteAlways)
}

// WithDownloadOverwriteNever sets download to never overwrite existing files
func WithDownloadOverwriteNever() DownloadOption {
	return WithDownloadOverwritePolicy(OverwriteNever)
}

// WithDownloadOverwriteIfNewer sets download to overwrite only if remote file is newer
func WithDownloadOverwriteIfNewer() DownloadOption {
	return WithDownloadOverwritePolicy(OverwriteIfNewer)
}

// WithDownloadOverwriteIfDifferentSize sets download to overwrite only if file sizes differ
func WithDownloadOverwriteIfDifferentSize() DownloadOption {
	return WithDownloadOverwritePolicy(OverwriteIfDifferentSize)
}

// WithDownloadOverwriteIfNewerOrDifferentSize sets download to overwrite if newer or different size
func WithDownloadOverwriteIfNewerOrDifferentSize() DownloadOption {
	return WithDownloadOverwritePolicy(OverwriteIfNewerOrDifferentSize)
}
