package sftp_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pkg/sftp"
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
				server, err := sftp.NewServer(channel, sftp.WithServerWorkingDirectory(s.tempDir))
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
