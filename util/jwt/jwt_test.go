package jwt_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	jwtutil "github.com/kittipat1413/go-common/util/jwt"
	"github.com/stretchr/testify/require"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	CustomField string `json:"customField"`
}

func TestJWTManager(t *testing.T) {
	t.Run("TestNewJWTManager", func(t *testing.T) {
		t.Run("MissingSigningMethod", func(t *testing.T) {
			mgr, err := jwtutil.NewJWTManager("", []byte("some-secret"))
			require.Error(t, err)
			require.Nil(t, mgr)
		})

		t.Run("MissingSigningKey", func(t *testing.T) {
			mgr, err := jwtutil.NewJWTManager(jwtutil.HS256, nil)
			require.Error(t, err)
			require.Nil(t, mgr)
		})

		t.Run("UnsupportedMethod", func(t *testing.T) {
			var unsupported jwtutil.SupportedSigningMethod = "ABC123"
			mgr, err := jwtutil.NewJWTManager(unsupported, []byte("some-secret"))
			require.Error(t, err)
			require.Nil(t, mgr)
		})

		t.Run("HS256 Success", func(t *testing.T) {
			mgr, err := jwtutil.NewJWTManager(jwtutil.HS256, []byte("mysecretkey"))
			require.NoError(t, err)
			require.NotNil(t, mgr)
		})

		t.Run("RS256 Success", func(t *testing.T) {
			mgr, err := jwtutil.NewJWTManager(jwtutil.RS256, mustGenRSAPrivateKeyPEM(t, 2048))
			require.NoError(t, err)
			require.NotNil(t, mgr)
		})
	})

	t.Run("TestCreateToken", func(t *testing.T) {
		t.Run("HS256 Success", func(t *testing.T) {
			hsManager, err := jwtutil.NewJWTManager(jwtutil.HS256, []byte("mysecretkey"))
			require.NoError(t, err)
			require.NotNil(t, hsManager)

			claims := &CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "test-issuer",
					Subject:   "test-subject",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
				},
				CustomField: "HS256CustomData",
			}

			tokenStr, err := hsManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)
			require.NotEmpty(t, tokenStr)
		})

		t.Run("RS256 Success", func(t *testing.T) {
			rsManager, err := jwtutil.NewJWTManager(jwtutil.RS256, mustGenRSAPrivateKeyPEM(t, 2048))
			require.NoError(t, err)
			require.NotNil(t, rsManager)

			claims := &jwt.RegisteredClaims{
				Issuer:    "test-issuer-rs256",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			}

			tokenStr, err := rsManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)
			require.NotEmpty(t, tokenStr)
		})

		t.Run("RS256 Invalid PEM", func(t *testing.T) {
			rsManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(`-----BEGIN RSA PRIVATE KEY-----
INVALID-PEM
-----END RSA PRIVATE KEY-----
`))
			require.NoError(t, err) // Manager is created, but real error occurs on create.

			claims := &jwt.RegisteredClaims{
				Issuer:    "test-issuer-rs256-bad",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			}

			tokenStr, err := rsManager.CreateToken(context.Background(), claims)
			require.Error(t, err)
			require.Empty(t, tokenStr)
		})
	})

	t.Run("TestParseAndValidateToken", func(t *testing.T) {
		// Setup HS256 manager
		hsManager, err := jwtutil.NewJWTManager(jwtutil.HS256, []byte("mysecretkey"))
		require.NoError(t, err)
		require.NotNil(t, hsManager)

		// Setup RS256 manager
		rsManager, err := jwtutil.NewJWTManager(jwtutil.RS256, mustGenRSAPrivateKeyPEM(t, 2048))
		require.NoError(t, err)
		require.NotNil(t, rsManager)

		t.Run("HS256 Valid Token", func(t *testing.T) {
			claims := &CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					Issuer:    "hs256-issuer",
				},
				CustomField: "HS256Data",
			}
			tokenStr, err := hsManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)

			parsedClaims := &CustomClaims{}
			err = hsManager.ParseAndValidateToken(context.Background(), tokenStr, parsedClaims)
			require.NoError(t, err)

			// Validate the claims
			require.Equal(t, claims.Issuer, parsedClaims.Issuer)
			require.Equal(t, claims.CustomField, parsedClaims.CustomField)
		})

		t.Run("HS256 Invalid Token (bad signature)", func(t *testing.T) {
			// Create a token with a different key
			otherManager, err := jwtutil.NewJWTManager(jwtutil.HS256, []byte("another-secret"))
			require.NoError(t, err)

			claims := &jwt.RegisteredClaims{Issuer: "hs256-issuer-bad"}
			badTokenStr, err := otherManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)

			parsedClaims := &jwt.RegisteredClaims{}
			err = hsManager.ParseAndValidateToken(context.Background(), badTokenStr, parsedClaims)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to parse token")
		})

		t.Run("HS256 Unexpected Signing Method", func(t *testing.T) {
			// Create a token using RS256
			rsClaims := &jwt.RegisteredClaims{Issuer: "rs256-issuer"}
			rsTokenStr, err := rsManager.CreateToken(context.Background(), rsClaims)
			require.NoError(t, err)

			parsedClaims := &jwt.RegisteredClaims{}
			err = hsManager.ParseAndValidateToken(context.Background(), rsTokenStr, parsedClaims)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unexpected signing method expected HS256 but got RS256")
		})

		t.Run("RS256 Valid Token", func(t *testing.T) {
			claims := &jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				Issuer:    "rs256-issuer",
			}
			tokenStr, err := rsManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)
			require.NotEmpty(t, tokenStr)

			parsedClaims := &jwt.RegisteredClaims{}
			err = rsManager.ParseAndValidateToken(context.Background(), tokenStr, parsedClaims)
			require.NoError(t, err)
			require.Equal(t, claims.Issuer, parsedClaims.Issuer)
		})

		t.Run("RS256 Invalid Token (bad signature)", func(t *testing.T) {
			// otherManager uses a different 2048-bit key so signature won't match rsManager's key
			otherKey := mustGenRSAPrivateKeyPEM(t, 2048)
			otherManager, err := jwtutil.NewJWTManager(jwtutil.RS256, otherKey)
			require.NoError(t, err) // Manager was created, actual error will be during parse

			// Create a valid token from HS256 manager
			claims := &jwt.RegisteredClaims{Issuer: "rs256-issuer-bad"}
			badTokenStr, err := otherManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)

			parsedClaims := &jwt.RegisteredClaims{}
			err = rsManager.ParseAndValidateToken(context.Background(), badTokenStr, parsedClaims)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to parse token")
		})

		t.Run("RS256 Unexpected Signing Method", func(t *testing.T) {
			// Create a valid token from HS256 manager
			claims := &jwt.RegisteredClaims{Issuer: "rs256-issuer-bad"}
			hsTokenStr, err := hsManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)

			parsedClaims := &jwt.RegisteredClaims{}
			err = rsManager.ParseAndValidateToken(context.Background(), hsTokenStr, parsedClaims)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unexpected signing method expected RS256 but got HS256")
		})

		t.Run("RS256 Invalid PEM when parsing", func(t *testing.T) {
			// Manager with invalid PEM
			rsInvalidManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(`-----BEGIN RSA PRIVATE KEY-----
INVALID-PEM
-----END RSA PRIVATE KEY-----
`))
			require.NoError(t, err) // Manager was created, actual error will be during parse

			claims := &jwt.RegisteredClaims{Issuer: "rs256-issuer-bad-pem"}
			// We still need a token. We'll create a "valid" token using the valid manager
			validTokenStr, err := rsManager.CreateToken(context.Background(), claims)
			require.NoError(t, err)

			parsedClaims := &jwt.RegisteredClaims{}
			err = rsInvalidManager.ParseAndValidateToken(context.Background(), validTokenStr, parsedClaims)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid RSA private key")
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
