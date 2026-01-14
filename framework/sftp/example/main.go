package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/kittipat1413/go-common/framework/sftp"
)

func main() {
	// Set up default logger configuration to DEBUG level with structured JSON formatter
	// for better visibility during this example.
	err := logger.SetDefaultLoggerConfig(logger.Config{Level: logger.DEBUG, Formatter: &logger.StructuredJSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     true,
	}})
	if err != nil {
		fmt.Println("Failed to set logger config:", err)
		return
	}

	// SFTP client configuration
	config := sftp.Config{
		Authentication: sftp.AuthConfig{
			Host:     "localhost", // hostname or IP address of the SFTP server
			Port:     22,          // port number of the SFTP server
			Username: "user",      // username for authentication

			// Password authentication
			// Method: sftp.AuthPassword, // authentication method: AuthPassword or AuthPrivateKey
			// Password: "password", // password for AuthPassword method

			// Private key authentication
			Method:         sftp.AuthPrivateKey,    // authentication method: AuthPassword or AuthPrivateKey
			PrivateKeyPath: "/path/to/private/key", // (optional) path to private key file for AuthPrivateKey method
			PrivateKeyData: []byte{},               // (optional) private key data for AuthPrivateKey method

			// HostKeyCallback: ssh.InsecureIgnoreHostKey(), // host key callback for server verification
		},
		// Connection: sftp.ConnectionConfig{}, // (optional) connection settings
		// Transfer:   sftp.TransferConfig{},   // (optional) transfer settings
	}

	// Create and use SFTP client
	client, err := sftp.NewClient(config)
	if err != nil {
		fmt.Println("Failed to create SFTP client:", err)
		return
	}
	defer func() {
		// Close all connections in the SFTP connection pool
		if cerr := client.Close(); cerr != nil {
			fmt.Println("Failed to close SFTP client:", cerr)
		}
	}()

	// Establishes a new SFTP connection and registers it in the connection pool
	err = client.Connect(context.Background())
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}

	// Upload a local file to the SFTP server with progress reporting and automatic directory creation
	err = client.Upload(context.Background(), "README.md", "/upload/remotefile.txt",
		sftp.WithUploadCreateDirs(true),
		sftp.WithUploadProgress(func(info sftp.ProgressInfo) {
			fmt.Printf("Uploaded %d/%d bytes\n", info.BytesTransferred, info.TotalBytes)
		}),
	)
	if err != nil {
		fmt.Println("Failed to upload file:", err)
		return
	}

	// Stat the uploaded file to get its information
	fileInfo, err := client.Stat(context.Background(), "/upload/remotefile.txt")
	if err != nil {
		fmt.Println("Failed to stat file:", err)
		return
	}
	fmt.Printf("Uploaded file size: %d bytes\n", fileInfo.Size())

	// Rename the uploaded file on the SFTP server
	err = client.Rename(context.Background(), "/upload/remotefile.txt", "/upload/remotefile_renamed.txt")
	if err != nil {
		fmt.Println("Failed to rename file:", err)
		return
	}

	// Download the renamed file back to local storage with progress reporting
	err = client.Download(context.Background(), "/upload/remotefile_renamed.txt", "downloaded_remotefile.txt",
		sftp.WithDownloadProgress(func(info sftp.ProgressInfo) {
			fmt.Printf("Downloaded %d/%d bytes\n", info.BytesTransferred, info.TotalBytes)
		}),
	)
	if err != nil {
		fmt.Println("Failed to download file:", err)
		return
	}

	// Clean up: remove the uploaded files and directory from the SFTP server
	err = client.Remove(context.Background(), "/upload/remotefile_renamed.txt")
	if err != nil {
		fmt.Println("Failed to remove directory:", err)
		return
	}
	err = client.Remove(context.Background(), "/upload")
	if err != nil {
		fmt.Println("Failed to remove directory:", err)
		return
	}

	// List files in the root directory of the SFTP server
	list, err := client.List(context.Background(), "/")
	if err != nil {
		fmt.Println("Failed to list directory:", err)
		return
	}
	for _, item := range list {
		fmt.Println(" -", item.Name())
	}
}
