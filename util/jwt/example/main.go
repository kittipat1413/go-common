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
	////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Example: HMAC-based token creation and validation
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

	////////////////////////////////////////////////////////////////////////////////////////////////////////////

	// Example: RSA-based token creation and validation
	const rsaPrivateKeyPEM = `
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
	manager, err = jwtutil.NewJWTManager(jwtutil.RS256, []byte(rsaPrivateKeyPEM))
	if err != nil {
		log.Fatalf("Failed to create JWTManager: %v", err)
	}

	// Prepare custom claims
	claims = &MyCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "example-RSA256",
			Subject:   "example-subject",
		},
		UserID: "abc123",
	}

	// Create the token
	tokenStringRS256, err := manager.CreateToken(ctx, claims)
	if err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}
	fmt.Println("Generated Token:", tokenStringRS256)

	// Validate the token
	parsedClaims = &MyCustomClaims{}
	err = manager.ParseAndValidateToken(ctx, tokenStringRS256, parsedClaims)
	if err != nil {
		log.Fatalf("Failed to validate token: %v", err)
	}

	fmt.Printf("Token is valid! UserID: %s, Issuer: %s\n", parsedClaims.UserID, parsedClaims.Issuer)
}
