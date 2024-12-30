package jwt

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockgen -source=./jwt.go -destination=./mocks/jwt.go -package=jwt_mocks
type JWTManager interface {
	// CreateToken generates a signed JWT token with the provided claims.
	// The claims should implement the jwt.Claims interface (e.g., *jwt.RegisteredClaims or a custom struct).
	CreateToken(ctx context.Context, claims jwt.Claims) (string, error)

	// ParseAndValidateToken parses and validates the token string, populating the provided claims struct if valid.
	// The user must pass a pointer to a claims struct (e.g., `&MyCustomClaims{}` or `&jwt.RegisteredClaims{}`)
	// that implements `jwt.Claims`. The function validates the token and populates the provided struct.
	ParseAndValidateToken(ctx context.Context, tokenString string, claims jwt.Claims) error
}

// SupportedSigningMethod defines the supported JWT signing methods for token creation and validation.
type SupportedSigningMethod string

const (
	HS256 SupportedSigningMethod = "HS256"
	RS256 SupportedSigningMethod = "RS256"
)

// getJwtSigningMethod maps the SupportedSigningMethod to the corresponding jwt.SigningMethod.
func (m SupportedSigningMethod) getJwtSigningMethod() (jwt.SigningMethod, error) {
	switch m {
	case HS256:
		return jwt.SigningMethodHS256, nil
	case RS256:
		return jwt.SigningMethodRS256, nil
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", m)
	}
}

type jwtManager struct {
	// signingMethod specifies the algorithm used for signing and validating JWT tokens (e.g., HS256, RS256).
	signingMethod jwt.SigningMethod

	// key contains the cryptographic key used for signing and verifying tokens:
	// - For HMAC-based algorithms (e.g., HS256), it is the shared secret key.
	// - For RSA-based algorithms (e.g., RS256), it is the PEM-encoded private key.
	signingKey []byte
}

// NewJWTManager initializes a new JWT manager with the given signing method and key.
//
// Params:
//   - signingMethod: The algorithm used for signing and validating JWT tokens (e.g., HS256, RS256).
//   - signingKey: The cryptographic key used for signing and verifying tokens (e.g., shared secret, PEM-encoded private key).
func NewJWTManager(signingMethod SupportedSigningMethod, signingKey []byte) (JWTManager, error) {
	if signingMethod == "" {
		return nil, errors.New("failed to create JWT manager: missing signing method")
	}
	if len(signingKey) == 0 {
		return nil, errors.New("failed to create JWT manager: missing signing key")
	}

	jwtSigningMethod, err := signingMethod.getJwtSigningMethod()
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	return &jwtManager{
		signingMethod: jwtSigningMethod,
		signingKey:    signingKey,
	}, nil
}

// CreateToken generates a signed JWT token with the provided claims.
// The claims should implement the jwt.Claims interface (e.g., *jwt.RegisteredClaims or a custom struct).
func (m *jwtManager) CreateToken(ctx context.Context, claims jwt.Claims) (string, error) {
	// Create a new token object with the desired signing method and claims.
	token := jwt.NewWithClaims(m.signingMethod, claims)

	// Sign the token using the configured method.
	switch m.signingMethod.(type) {
	case *jwt.SigningMethodHMAC:
		// HMAC: signingKey is the shared secret.
		return token.SignedString(m.signingKey)
	case *jwt.SigningMethodRSA:
		// RSA: signingKey must be the PEM-encoded private key.
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(m.signingKey)
		if err != nil {
			return "", fmt.Errorf("invalid RSA private key: %w", err)
		}
		return token.SignedString(privateKey)
	default:
		return "", fmt.Errorf("unsupported signing method for token creation: %v", m.signingMethod.Alg())
	}
}

// ParseAndValidateToken parses and validates the token string, populating the provided claims struct if valid.
// If the token is invalid or the claims cannot be validated, an error is returned.
func (m *jwtManager) ParseAndValidateToken(ctx context.Context, tokenString string, claims jwt.Claims) error {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method matches the configured one.
		if token.Method.Alg() != m.signingMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method expected %s but got %s", m.signingMethod.Alg(), token.Method.Alg())
		}

		switch m.signingMethod.(type) {
		case *jwt.SigningMethodHMAC:
			// HMAC: use the shared secret to verify signature.
			return m.signingKey, nil
		case *jwt.SigningMethodRSA:
			// RSA: parse the public key for verification.
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(m.signingKey)
			if err != nil {
				return nil, fmt.Errorf("invalid RSA private key: %w", err)
			}
			publicKey := &privateKey.PublicKey
			return publicKey, nil
		default:
			return nil, fmt.Errorf("unsupported signing method for token validation: %v", m.signingMethod.Alg())
		}
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}
	if !parsedToken.Valid {
		return errors.New("invalid token: token is not valid")
	}
	return nil
}
