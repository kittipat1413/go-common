[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# SFTP Package
The SFTP package provides a SFTP client with connection pooling, flexible authentication, and advanced file transfer capabilities. Built for reliability and ease of use, it simplifies secure file operations in distributed systems.

## Features
- **Unified SFTP Client Interface**
  - Complete file operations: `Upload`, `Download`, `List`, `Mkdir`, `Remove`, `Rename`, `Stat`
  - Context-aware operations with cancellation support
- **Connection Pooling**
  - Efficient connection reuse with automatic management
  - Configurable pool size and idle timeout
  - Health checks and automatic cleanup
  - Thread-safe concurrent access
- **Flexible Authentication**
  - Password authentication
  - Private key authentication (file or in-memory)
  - Passphrase-protected private keys
  - Extensible authentication handlers
- **Advanced File Transfer**
  - Configurable buffer size for optimal performance
  - Real-time progress callbacks
  - Smart overwrite policies (always, never, if newer, if different size)
  - Automatic directory creation
  - Permission preservation
- **Reliability & Retry**
  - Built-in retry mechanism with exponential backoff
  - Configurable retry policies
  - Connection health monitoring
  - Graceful error handling


## Installation
```bash
go get github.com/kittipat1413/go-common/framework/sftp
```

## Documentation
[![Go Reference](https://pkg.go.dev/badge/github.com/kittipat1413/go-common/framework/sftp.svg)](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/sftp)

For detailed API documentation, examples, and usage patterns, visit the [Go Package Documentation](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/sftp).

## Usage

### Client Interface
The core of the package is the `Client` interface:
```go
type Client interface {
    Connect(ctx context.Context) error
    Close() error
    
    Upload(ctx context.Context, localPath, remotePath string, opts ...UploadOption) error
    Download(ctx context.Context, remotePath, localPath string, opts ...DownloadOption) error
    List(ctx context.Context, remotePath string) ([]os.FileInfo, error)
    Mkdir(ctx context.Context, remotePath string) error
    Remove(ctx context.Context, remotePath string) error
    Rename(ctx context.Context, oldPath, newPath string) error
    Stat(ctx context.Context, remotePath string) (os.FileInfo, error)
}
```

### 🚀 Quick Start

#### Creating an SFTP Client
**Password Authentication:**
```go
import (
    "context"
    "github.com/kittipat1413/go-common/framework/sftp"
)

func main() {
    config := sftp.Config{
        Authentication: sftp.AuthConfig{
            Host:     "sftp.example.com",
            Port:     22,
            Username: "myuser",
            Method:   sftp.AuthPassword,
            Password: "mypassword",
        },
    }

    client, err := sftp.NewClient(config)
    if err != nil {
        log.Fatalf("Failed to create SFTP client: %v", err)
    }
    defer client.Close()

    // Test connection
    if err := client.Connect(context.Background()); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
}
```

**Private Key Authentication:**
```go
config := sftp.Config{
    Authentication: sftp.AuthConfig{
        Host:           "sftp.example.com",
        Port:           22,
        Username:       "myuser",
        Method:         sftp.AuthPrivateKey,
        PrivateKeyPath: "/path/to/private/key",
        // Or use in-memory key data:
        // PrivateKeyData: []byte("-----BEGIN RSA PRIVATE KEY-----\n..."),
    },
}

client, err := sftp.NewClient(config)
```

### 📤 Uploading Files
**Basic Upload:**
```go
ctx := context.Background()
err := client.Upload(ctx, "local/file.txt", "/remote/file.txt")
if err != nil {
    log.Fatalf("Upload failed: %v", err)
}
```

**Upload with Options:**
```go
err := client.Upload(ctx, "local/report.pdf", "/reports/2024/report.pdf",
    sftp.WithUploadCreateDirs(true),           // Create parent directories
    sftp.WithUploadPreservePermissions(true),  // Preserve file permissions
    sftp.WithUploadOverwriteIfNewer(),         // Only overwrite if newer
    sftp.WithUploadProgress(func(info sftp.ProgressInfo) {
        progress := float64(info.BytesTransferred) / float64(info.TotalBytes) * 100
        fmt.Printf("Upload progress: %.2f%%\n", progress)
    }),
)
```

### 📥 Downloading Files
**Basic Download:**
```go
err := client.Download(ctx, "/remote/file.txt", "local/file.txt")
if err != nil {
    log.Fatalf("Download failed: %v", err)
}
```

**Download with Options:**
```go
err := client.Download(ctx, "/remote/large-file.zip", "local/large-file.zip",
    sftp.WithDownloadCreateDirs(true),
    sftp.WithDownloadPreservePermissions(true),
    sftp.WithDownloadOverwriteIfNewerOrDifferentSize(),
    sftp.WithDownloadProgress(func(info sftp.ProgressInfo) {
        fmt.Printf("Downloaded %d/%d bytes\n", info.BytesTransferred, info.TotalBytes)
    }),
)
```

### 📁 Directory Operations
```go
// Create directory (with parents)
err := client.Mkdir(ctx, "/remote/path/to/new/dir")

// List directory contents
files, err := client.List(ctx, "/remote/path")
for _, file := range files {
    fmt.Printf("%s - %d bytes\n", file.Name(), file.Size())
}

// Remove file or directory (recursive)
err := client.Remove(ctx, "/remote/path/to/file")

// Rename or move file
err := client.Rename(ctx, "/old/path", "/new/path")

// Get file info
info, err := client.Stat(ctx, "/remote/file.txt")
fmt.Printf("Size: %d, Modified: %v\n", info.Size(), info.ModTime())
```

## Configuration
SFTP client behavior can be customized via the `Config` struct:
```go
config := sftp.Config{
    Authentication: sftp.AuthConfig{
        Host:            "sftp.example.com",
        Port:            22,
        Username:        "myuser",
        Method:          sftp.AuthPrivateKey,
        PrivateKeyPath:  "/path/to/key",
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    },
    Connection: sftp.ConnectionConfig{
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
    Transfer: sftp.TransferConfig{
        BufferSize:          32 * 1024, // 32KB
        CreateDirs:          true,
        PreservePermissions: false,
    },
}
```

### Configuration Options

#### AuthConfig
| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Host` | `string` | SFTP server hostname or IP | Required |
| `Port` | `int` | SFTP server port | `22` |
| `Username` | `string` | Authentication username | Required |
| `Method` | `AuthMethod` | Auth method (`AuthPassword` or `AuthPrivateKey`) | Required |
| `Password` | `string` | Password (for `AuthPassword`) | - |
| `PrivateKeyPath` | `string` | Path to private key file | - |
| `PrivateKeyData` | `[]byte` | Private key data (in-memory) | - |
| `HostKeyCallback` | `ssh.HostKeyCallback` | Server verification callback | Insecure (accept all) |

#### ConnectionConfig
| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Timeout` | `time.Duration` | Connection timeout | `30s` |
| `MaxConnections` | `int` | Max connections in pool | `10` |
| `IdleTimeout` | `time.Duration` | Idle connection timeout | `5m` |
| `RetryPolicy` | `retry.Config` | Retry configuration | Exponential backoff |

#### TransferConfig
| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `BufferSize` | `int` | Transfer buffer size (bytes) | `32768` (32KB) |
| `CreateDirs` | `bool` | Auto-create directories | `false` |
| `PreservePermissions` | `bool` | Preserve file permissions | `false` |
| `ProgressCallback` | `ProgressCallback` | Progress reporting function | `nil` |

### Overwrite Policies
Control how the client handles existing files:

- `OverwriteAlways` - Always overwrite existing files (default)
- `OverwriteNever` - Never overwrite, return error if exists
- `OverwriteIfNewer` - Overwrite only if source is newer
- `OverwriteIfDifferentSize` - Overwrite only if sizes differ
- `OverwriteIfNewerOrDifferentSize` - Overwrite if newer OR different size

```go
// Upload only if local file is newer
client.Upload(ctx, "local.txt", "remote.txt", 
    sftp.WithUploadOverwriteIfNewer())

// Download only if remote file is newer
client.Download(ctx, "remote.txt", "local.txt",
    sftp.WithDownloadOverwriteIfNewer())
```

### Connection Pool Management
Connections are automatically managed
The pool will:
- Reuse existing connections when available
- Create new connections up to MaxConnections limit
- Automatically close idle connections after IdleTimeout
- Perform health checks before returning connections
```go

// Manually close all connections (e.g., during shutdown)
defer client.Close()
```

## Error Handling
### Common Errors
- `ErrConfiguration` – Invalid configuration
- `ErrAuthentication` – Authentication failure
- `ErrConnection` – Connection or network issues
- `ErrConnectionPoolFull` – No available connections in the pool
- `ErrConnectionPoolClosed` – Connection pool is closed
- `ErrFileNotFound` – File does not exist, either remotely or locally
- `ErrDataTransfer` – Errors during data transfer, including overwrite policy violations

## Example
You can find a complete working example in the repository under [framework/sftp/example](example/).
