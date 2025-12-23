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
		t.Run("Valid Credentials", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("testuser", "testpass")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods for valid credentials
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("Empty Username", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("", "testpass")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Empty Password", func(t *testing.T) {
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

		t.Run("Valid Key Data", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", validPrivateKey, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods for valid credentials
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("Valid Key Data With Passphrase", func(t *testing.T) {
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

		t.Run("Empty Username", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("", validPrivateKey, "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Valid Key Data With Incorrect Passphrase", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", validPrivateKey, "testpassphrase")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods for valid credentials
			_, err = handler.GetAuthMethods()
			assert.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Empty Key Data", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", []byte{}, "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Invalid Key Format", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", []byte("invalid key data"), "")
			// ValidateCredentials only checks username and key existence, not format
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Invalid PEM Format", func(t *testing.T) {
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

		t.Run("Valid Key File Path", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", validPrivateKey)

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods reads and parses the key file successfully
			methods, err := handler.GetAuthMethods()
			assert.NoError(t, err)
			assert.Len(t, methods, 1)
		})

		t.Run("Valid Key File Path With Passphrase", func(t *testing.T) {
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

		t.Run("Empty Username", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", validPrivateKey)

			handler := sftp.NewPrivateKeyAuthHandler("", tmpFile, "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Valid Key File Path With Incorrect Passphrase", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", validPrivateKey)

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "testpassphrase")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// Test GetAuthMethods reads and parses the key file successfully
			_, err = handler.GetAuthMethods()
			assert.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Empty Key File Path", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandler("testuser", "", "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Non-Existent Key File", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandler("testuser", "/non/existent/path/key.pem", "")
			err := handler.ValidateCredentials()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Invalid Key File Content", func(t *testing.T) {
			tmpFile := mustWriteTempFile(t, "test-key.pem", []byte("invalid key content"))

			handler := sftp.NewPrivateKeyAuthHandler("testuser", tmpFile, "")
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Invalid PEM File Content", func(t *testing.T) {
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

		t.Run("Password Auth Missing Host", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Username: "testuser",
				Method:   sftp.AuthPassword,
				Password: "testpass",
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrConfiguration)
		})

		t.Run("Password Auth Missing Username", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Method:   sftp.AuthPassword,
				Password: "testpass",
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrConfiguration)
		})

		t.Run("Password Auth", func(t *testing.T) {
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

		t.Run("Password Auth Missing Password", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthPassword,
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Private Key Auth Missing Key", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthPrivateKey,
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Nil(t, handler)
			require.ErrorIs(t, err, sftp.ErrAuthentication)
		})

		t.Run("Private Key Auth With Key Data", func(t *testing.T) {
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

		t.Run("Private Key Auth With Key Path", func(t *testing.T) {
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

		t.Run("Unsupported Auth Method", func(t *testing.T) {
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
