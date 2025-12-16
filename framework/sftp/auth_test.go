package sftp_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/kittipat1413/go-common/framework/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})

		t.Run("Empty Password", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("testuser", "")
			err := handler.ValidateCredentials()
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})

		t.Run("Whitespace Username", func(t *testing.T) {
			handler := sftp.NewPasswordAuthHandler("   ", "testpass")
			err := handler.ValidateCredentials()
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})
	})
}

// TestPrivateKeyAuthHandler tests private key authentication functionality
func TestPrivateKeyAuthHandler(t *testing.T) {
	t.Run("TestNewPrivateKeyAuthHandler", func(t *testing.T) {
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
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", validPrivateKey, "testpassphrase")
			err := handler.ValidateCredentials()
			require.NoError(t, err)
		})

		t.Run("Empty Username", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("", validPrivateKey, "")
			err := handler.ValidateCredentials()
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})

		t.Run("Whitespace Username", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("   ", validPrivateKey, "")
			err := handler.ValidateCredentials()
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})

		t.Run("No Key Data Or Path", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandler("testuser", "", "")
			err := handler.ValidateCredentials()
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})

		t.Run("Invalid Key Format", func(t *testing.T) {
			handler := sftp.NewPrivateKeyAuthHandlerWithData("testuser", []byte("invalid key data"), "")
			// ValidateCredentials only checks username and key existence, not format
			err := handler.ValidateCredentials()
			require.NoError(t, err)

			// The actual error occurs when getting auth methods
			_, err = handler.GetAuthMethods()
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
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
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})
	})
}

// TestCreateAuthHandler tests the auth handler factory function
func TestCreateAuthHandler(t *testing.T) {
	t.Run("TestCreateAuthHandlerFactory", func(t *testing.T) {
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
			require.Error(t, err)
			require.Nil(t, handler)
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
		})

		t.Run("Private Key Auth Missing Key", func(t *testing.T) {
			authConfig := sftp.AuthConfig{
				Host:     "localhost",
				Username: "testuser",
				Method:   sftp.AuthPrivateKey,
			}
			handler, err := sftp.CreateAuthHandler(authConfig)
			require.Error(t, err)
			require.Nil(t, handler)
			require.ErrorAs(t, err, &sftp.ErrAuthentication)
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
