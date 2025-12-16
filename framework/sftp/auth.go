package sftp

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

//go:generate mockgen -source=./auth.go -destination=./mocks/auth.go -package=sftp_mocks

// AuthenticationHandler manages different authentication methods
type AuthenticationHandler interface {
	GetAuthMethods() ([]ssh.AuthMethod, error)
	ValidateCredentials() error
}

// AuthMethod defines the type of authentication to use
type AuthMethod int

const (
	AuthPassword AuthMethod = iota
	AuthPrivateKey
)

// CreateAuthHandler creates an appropriate authentication handler based on the auth config
func CreateAuthHandler(authConfig AuthConfig) (AuthenticationHandler, error) {
	// Merge with default config to ensure all fields are set
	mergedConfig := mergeAuthConfig(DefaultConfig().Authentication, authConfig)
	if err := validateAuthConfig(mergedConfig); err != nil {
		return nil, err // errors are wrapped in validateConfig
	}

	// Create the appropriate authentication handler
	switch mergedConfig.Method {
	case AuthPassword:
		if mergedConfig.Password == "" {
			return nil, fmt.Errorf("%w: password is required for password authentication", ErrAuthentication)
		}
		return NewPasswordAuthHandler(mergedConfig.Username, mergedConfig.Password), nil

	case AuthPrivateKey:
		if len(mergedConfig.PrivateKeyData) == 0 && mergedConfig.PrivateKeyPath == "" {
			return nil, fmt.Errorf("%w: private key path or data is required for private key authentication", ErrAuthentication)
		}
		if len(mergedConfig.PrivateKeyData) > 0 {
			return NewPrivateKeyAuthHandlerWithData(mergedConfig.Username, mergedConfig.PrivateKeyData, ""), nil
		} else {
			return NewPrivateKeyAuthHandler(mergedConfig.Username, mergedConfig.PrivateKeyPath, ""), nil
		}
	default:
		return nil, fmt.Errorf("%w: unsupported authentication method", ErrAuthentication)
	}
}

// PasswordAuthHandler handles password-based authentication
type PasswordAuthHandler struct {
	username string
	password string
}

// NewPasswordAuthHandler creates a new password authentication handler
func NewPasswordAuthHandler(username, password string) *PasswordAuthHandler {
	return &PasswordAuthHandler{
		username: username,
		password: password,
	}
}

// GetAuthMethods returns SSH authentication methods for password auth
func (p *PasswordAuthHandler) GetAuthMethods() ([]ssh.AuthMethod, error) {
	if err := p.ValidateCredentials(); err != nil {
		return nil, err
	}

	authMethod := ssh.Password(p.password)
	return []ssh.AuthMethod{authMethod}, nil
}

// ValidateCredentials validates password authentication credentials
func (p *PasswordAuthHandler) ValidateCredentials() error {
	if strings.TrimSpace(p.username) == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrAuthentication)
	}
	if strings.TrimSpace(p.password) == "" {
		return fmt.Errorf("%w: password cannot be empty", ErrAuthentication)
	}
	return nil
}

// PrivateKeyAuthHandler handles SSH private key authentication
type PrivateKeyAuthHandler struct {
	username       string
	privateKeyPath string
	privateKeyData []byte
	passphrase     string
}

// NewPrivateKeyAuthHandler creates a new private key authentication handler
func NewPrivateKeyAuthHandler(username, privateKeyPath string, passphrase string) *PrivateKeyAuthHandler {
	return &PrivateKeyAuthHandler{
		username:       username,
		privateKeyPath: privateKeyPath,
		passphrase:     passphrase,
	}
}

// NewPrivateKeyAuthHandlerWithData creates a new private key authentication handler with key data
func NewPrivateKeyAuthHandlerWithData(username string, privateKeyData []byte, passphrase string) *PrivateKeyAuthHandler {
	return &PrivateKeyAuthHandler{
		username:       username,
		privateKeyData: privateKeyData,
		passphrase:     passphrase,
	}
}

// GetAuthMethods returns SSH authentication methods for private key auth
func (k *PrivateKeyAuthHandler) GetAuthMethods() ([]ssh.AuthMethod, error) {
	if err := k.ValidateCredentials(); err != nil {
		return nil, err
	}

	var keyData []byte
	var err error

	// Use provided key data or read from file
	if len(k.privateKeyData) > 0 {
		keyData = k.privateKeyData
	} else {
		keyData, err = os.ReadFile(k.privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to read private key file: %v", ErrAuthentication, err)
		}
	}

	// Parse the private key
	var signer ssh.Signer
	if k.passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(k.passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyData)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse private key: %v", ErrAuthentication, err)
	}

	authMethod := ssh.PublicKeys(signer)
	return []ssh.AuthMethod{authMethod}, nil
}

// ValidateCredentials validates private key authentication credentials
func (k *PrivateKeyAuthHandler) ValidateCredentials() error {
	if strings.TrimSpace(k.username) == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrAuthentication)
	}

	// Check if we have either key data or key path
	if len(k.privateKeyData) == 0 && strings.TrimSpace(k.privateKeyPath) == "" {
		return fmt.Errorf("%w: either private key data or private key path must be provided", ErrAuthentication)
	}

	// If using key path, validate file exists and is readable
	if len(k.privateKeyData) == 0 {
		if _, err := os.Stat(k.privateKeyPath); err != nil {
			return fmt.Errorf("%w: private key file not accessible: %v", ErrAuthentication, err)
		}
	}

	return nil
}
