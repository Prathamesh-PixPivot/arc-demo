package licensing

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

type Verifier struct {
	publicKey *rsa.PublicKey
}

// NewVerifier creates a new license verifier with the given public key
func NewVerifier(publicKeyPEM []byte) (*Verifier, error) {
	key, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	return &Verifier{publicKey: key}, nil
}

// NewVerifierFromFile loads the public key from a file
func NewVerifierFromFile(path string) (*Verifier, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}
	return NewVerifier(pemBytes)
}

// Verify parses and verifies a license string (JWT)
func (v *Verifier) Verify(licenseString string) (*License, error) {
	token, err := jwt.ParseWithClaims(licenseString, &LicenseClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid license: %w", err)
	}

	if claims, ok := token.Claims.(*LicenseClaims); ok && token.Valid {
		return &claims.License, nil
	}

	return nil, fmt.Errorf("invalid license claims")
}

// LicenseClaims wraps License to implement jwt.Claims
type LicenseClaims struct {
	License
	jwt.RegisteredClaims
}
