package sftp_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/kittipat1413/go-common/framework/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestPasswordAuthHandler tests password authentication functionality
func TestPasswordAuthHandler(t *testing.T) {
	t.Run("TestNewPasswordAuthHandler", func(t *testing.T) {
		t.Run("should create handler with valid username and password", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("testuser", "testpass")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods for valid credentials
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("should return error for empty username", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("", "testpass")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for empty password", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("testuser", "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})
	})
}

// TestPrivateKeyAuthHandler tests private key authentication functionality
func TestPrivateKeyAuthHandler(t *testing.T) {
	t.Run("NewPrivateKeyAuthHandlerWithData", func(t *testing.T) {
		validPrivateKey := mustGenRSAPrivateKeyPEM(t, 2048)

		t.Run("should create handler with valid key data", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", validPrivateKey, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods for valid credentials
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("should create handler with valid key data and passphrase", func(t *testing.T) {
			passphrase := "testpassphrase"
			encryptedKey := mustGenEncryptedRSAPrivateKeyPEM(t, 2048, passphrase)

			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", encryptedKey, passphrase)
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods successfully parses encrypted key with correct passphrase
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("should return error for empty username", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("", validPrivateKey, "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for valid key data with incorrect passphrase", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", validPrivateKey, "testpassphrase")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods for valid credentials
			_, err = handler.GetAuthMethods()
			assert.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for empty key data", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", []byte{}, "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for invalid key format", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", []byte("invalid key data"), "")
			// ValidateCredentials only checks username and key existence, not format
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for invalid PEM format", func(t *testing.T) {
			invalidPEM := []byte(`-----BEGIN RSA PRIVATE KEY-----
	INVALID-PEM-DATA
	-----END RSA PRIVATE KEY-----`)
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", invalidPEM, "")
			// ValidateCredentials only checks username and key existence, not format
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})
	})

	t.Run("TestPrivateKeyPath", func(t *testing.T) {
		validPrivateKey := mustGenRSAPrivateKeyPEM(t, 2048)

		t.Run("should create handler with valid key file path", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", validPrivateKey)

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods reads and parses the key file successfully
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("should create handler with valid key file path and passphrase", func(t *testing.T) {
			passphrase := "testpassphrase"
			encryptedKey := mustGenEncryptedRSAPrivateKeyPEM(t, 2048, passphrase)
			tmpFile := mustWriteTempFile(t, "test-key-encrypted.pem", encryptedKey)

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, passphrase)
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods reads and parses encrypted key file with correct passphrase
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("should return error for empty username", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", validPrivateKey)

			handler := sftp.NewPrivateKeyAuthHandler("", tmpFile, "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for valid key file path with incorrect passphrase", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", validPrivateKey)

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "testpassphrase")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods reads and parses the key file successfully
			_, err = handler.GetAuthMethods()
			assert.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for empty key file path", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandler("testuser", "", "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for non-existent key file", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandler("testuser", "/non/existent/path/key.pem", "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for invalid key file content", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", []byte("invalid key content"))

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for invalid PEM file content", func(t *testing.T) {
			invalidPEM := []byte(`-----BEGIN RSA PRIVATE KEY-----
		INVALID-PEM-DATA
		-----END RSA PRIVATE KEY-----`)
			tmpFile := mustWriteTempFile(t, "test-key.pem", invalidPEM)

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})
	})
}

// TestCreateAuthHandler tests the auth handler factory function
func TestCreateAuthHandler(t *testing.T) {
	t.Run("TestCreateAuthHandlerFactory", func(t *testing.T) {

		t.Run("should return error for password auth missing host", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Username: "testuser",
				Method:   sftp.AuthPassword,
				Password: "testpass",
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrConfiguration)
		})

		t.Run("should return error for password auth missing username", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Method:   sftp.AuthPassword,
				Password: "testpass",
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrConfiguration)
		})

		t.Run("should create handler for password auth", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthPassword,
				Password: "testpass",
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.NoError(t, err)
			require.NotNil(t, handler)
		})

		t.Run("should return error for password auth missing password", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthPassword,
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should return error for private key auth missing key", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthPrivateKey,
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("should create handler for private key auth with key data", func(t *testing.T) {
			validPrivateKey := mustGenRSAPrivateKeyPEM(t, 2048)
			authConfig := sftp.AuthConfig{
				Host:           "localhost",
				Username:       "testuser",
				Method:         sftp.AuthPrivateKey,
				PrivateKeyData: validPrivateKey,
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.NoError(t, err)
			require.NotNil(t, handler)
		})

		t.Run("should create handler for private key auth with key path", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:           "localhost",
				Username:       "testuser",
				Method:         sftp.AuthPrivateKey,
				PrivateKeyPath: "/path/to/key",
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.NoError(t, err)
			require.NotNil(t, handler)
		})

		t.Run("should return error for unsupported auth method", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthMethod(999), // Invalid method
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})
	})
}

// mustGenRSAPrivateKeyPEM returns a PEM-encoded RSA PRIVATE KEY with the given size.
func mustGenRSAPrivateKeyPEM(t *testing.T, bits int) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
}

// mustGenEncryptedRSAPrivateKeyPEM returns a passphrase-encrypted PEM-encoded RSA PRIVATE KEY.
func mustGenEncryptedRSAPrivateKeyPEM(t *testing.T, bits int, passphrase string) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)

	// Use the modern ssh package to encrypt the key with passphrase
	encryptedKey, err := ssh.MarshalPrivateKeyWithPassphrase(key, "", []byte(passphrase))
	require.NoError(t, err)

	return pem.EncodeToMemory(encryptedKey)
}

// mustWriteTempFile creates a temporary file with the given content and returns its path.
// The file is automatically cleaned up when the test completes.
func mustWriteTempFile(t *testing.T, filename string, content []byte) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := tmpDir + "/" + filename

	err := os.WriteFile(filePath, content, 0600)
	require.NoError(t, err)

	return filePath
}
