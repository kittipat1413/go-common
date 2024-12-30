package jwt_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	jwtutil "github.com/kittipat1413/go-common/util/jwt"
	"github.com/stretchr/testify/require"
)

const validRSAPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAJLUdIexw37HyzB5VUoi3pIqbVifyV/
X6hl5DEY1vDRJFPpYSDngSPbwoBhzlhcYIMV0obyABs29AHDYd+rwDYWji62XaHGYBDGnKbhbDI1DMT
9ynkd0x1coxA4xTO+v1S8WJvp02w6TB5trokLOayhKizHkNynqerMbM2JqV09rAgMBAAECgYAwxF273
//lcOh8vh/k0rYH6A2PXOreaXE4aqr3+sr6trc/+uhqSKMTWZJi7KkSHJJt4rIBUKhx1u95i3wwzPBA
SmBipwl3ScP/HqeGnFnwqh6YrPmdPH3mptUEzO8wf1WldJS6o60i4b62nGU9UAvS4iYFjYSUN37Y5dP
S2+f5gQJBAOxm5hSVX4ppN0JSUDt4gr+P3RdmeQeBhw3F+OxIUI2TDQnzxAzo4BsGJovCYrTujrDoau
4y/SK/7WZYFT6zvKsCQQCfAJc0No1mI2T2t2DtV0WuxDLt96Xlv/7VsHQcODMvl/Fy5/ClsYlz4eLTe
vJTXFEtPI4FIKa3cKB2pQdKTThBAkA/CPUCug2+t21/prk0El8yuyal7bIJ+VTMrGRChMnN5k8Mv04g
bxwKuKogjBWLzyyHKYIRv9DVqj2gE46eqIh/AkAUnW39PgltMa+YcUQm4YbOVu/HfLFMrWzr5bnYIs0
4IXoTjNDdmrwYgzP2eV1Lw49ezxgWwBn9dKPJXjIoxwRBAkEAqeNVOd8gQDo3ZnLuWSDkT5a/8g2VaW
jrgnJmLKDLFWYjIpXQ1TxfluFSDQRPW4yzcWEULI2Jk0uGPG+NeAIMUg==
-----END RSA PRIVATE KEY-----
`
const invalidRSAPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
INVALID-PEM
-----END RSA PRIVATE KEY-----
`

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
			mgr, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(validRSAPrivateKey))
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
			rsManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(validRSAPrivateKey))
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
			rsManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(invalidRSAPrivateKey))
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
		rsManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(validRSAPrivateKey))
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
			otherManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBVQIBADANBgkqhkiG9w0BAQEFAASCAT8wggE7AgEAAkEAsCqGqHVfhADRMI05S2Lcl4VH7SFsnuq/fVbv1qW+N57ObX6jfxkWes1/7eYr
EvxD5fJuIJeiSxG6jEG0+Dc2JwIDAQABAkANVjLalwQzqKItpEtlnybnG7J9y81+3HPBx+6hV+vmJvH1rkGrVhRa80XmwVCr8GdG8VCpFsaS
1Jlk+d+aKWwFAiEA3tO+ziwIF8GI+9jkgWhBc4D345xd8wy9iXjUGTOGJWMCIQDKZHbHfdp/lh+N27jqErHF+s/+aLOTJK0VDf8idhT5bQIh
AKLkPhbv31amd2JMcvca5MXwIMb2V0PHK4OkncByhv0rAiEAo252v+av3uEh/9JSwqlv5lf/RwfTIlm2bk8MHA7QJw0CIDP+JTqPFs2P7RSG
o8LHst+W0b6dArnxPdzILlVMFbSD
-----END RSA PRIVATE KEY-----`))
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
			rsInvalidManager, err := jwtutil.NewJWTManager(jwtutil.RS256, []byte(invalidRSAPrivateKey))
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
