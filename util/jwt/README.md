# JWT Manager
JWT Manager provides a reusable and extensible interface for creating and validating JSON Web Tokens (JWTs) in Go, supporting both HMAC (HS256) and RSA (RS256) signing methods.

## Features
- **Token Creation**: Generate signed JWT tokens with custom claims.
- **Token Validation**: Parse and validate tokens with support for custom claims.
- **Pluggable Signing Methods**: Easily switch between HS256 (HMAC) and RS256 (RSA).

## Usage
### JWTManager Interface
```golang
type JWTManager interface {
    CreateToken(ctx context.Context, claims jwt.Claims) (string, error)
    ParseAndValidateToken(ctx context.Context, tokenString string, claims jwt.Claims) error
}
```
- **CreateToken**: Generates a signed JWT token with the provided claims.
  - _Params_:
    - `ctx`: Context for request tracing or cancellation.
	- `claims`: Claims to include in the token (must implement jwt.Claims).
  - _Returns_: Signed token string or an error.
- **ParseAndValidateToken**: Parses and validates a JWT token, populating the provided claims struct.
  - _Params_: 
    - `ctx`: Context for request tracing or cancellation.
    - `tokenString`: The JWT token string to validate.
    - `claims`: Pointer to a claims struct to populate (must implement jwt.Claims).
  - _Returns_: Error if validation fails; otherwise, populates the provided claims struct.

## Examples
```golang
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	jwtutil "github.com/kittipat1413/go-common/util/jwt"
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
}

func main() {
	ctx := context.Background()
	signingKey := []byte("super-secret-key")
	manager, err := jwtutil.NewJWTManager(jwtutil.HS256, signingKey)
	if err != nil {
		log.Fatalf("Failed to create JWTManager: %v", err)
	}

	// Prepare custom claims
	claims := &MyCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "example-HS256",
			Subject:   "example-subject",
		},
		UserID: "abc123",
	}

	// Create the token
	tokenStringHS256, err := manager.CreateToken(ctx, claims)
	if err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}
	fmt.Println("Generated Token:", tokenStringHS256)

	// Validate the token
	parsedClaims := &MyCustomClaims{}
	err = manager.ParseAndValidateToken(ctx, tokenStringHS256, parsedClaims)
	if err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}

	fmt.Printf("Token is valid! UserID: %s, Issuer: %s\n", parsedClaims.UserID, parsedClaims.Issuer)
}
```
> You can find a complete working example in the repository under [util/jwt/example](example/).