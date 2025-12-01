package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"time"

	"pixpivot/arc/internal/licensing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func main() {
	genKeys := flag.Bool("gen-keys", false, "Generate RSA key pair")
	signLicense := flag.Bool("sign", false, "Sign a license payload")
	privateKeyPath := flag.String("key", "private.pem", "Path to private key")
	output := flag.String("out", "license.lic", "Output file for license")

	flag.Parse()

	if *genKeys {
		generateKeys()
		return
	}

	if *signLicense {
		createAndSignLicense(*privateKeyPath, *output)
		return
	}

	fmt.Println("Usage: license-gen -gen-keys OR -sign -key private.pem")
}

func generateKeys() {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	// Save Private Key
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	os.WriteFile("private.pem", privPEM, 0600)
	fmt.Println("Generated private.pem")

	// Save Public Key
	pubBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	})
	os.WriteFile("public.pem", pubPEM, 0644)
	fmt.Println("Generated public.pem")
}

func createAndSignLicense(keyPath, outPath string) {
	// Load Private Key
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		panic(err)
	}

	// Create License Payload
	now := time.Now()
	expires := now.AddDate(1, 0, 0) // 1 year

	license := licensing.License{
		LicenseID:    uuid.New(),
		CustomerID:   uuid.New(),
		CustomerName: "Demo Customer",
		Issuer:       "PixPivot",
		IssuedAt:     now,
		ExpiresAt:    &expires,
		Type:         licensing.LicenseTypeSaaS,
		PlanTier:     licensing.PlanTierEnterprise,
		Features:     []string{"SSO", "AI_SCANNER"},
		Limits: licensing.LicenseLimits{
			MonthlyAPIRequests:  1000000,
			PIIRecordsProcessed: 50000,
			MaxUsers:            100,
			MaxDomains:          10,
		},
	}

	claims := licensing.LicenseClaims{
		License: license,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "PixPivot",
			Subject:   license.CustomerID.String(),
			Audience:  []string{"arc-backend"},
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        license.LicenseID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	os.WriteFile(outPath, []byte(signedString), 0644)
	fmt.Printf("Generated signed license at %s\n", outPath)
	fmt.Println("License Payload:")
	jsonBytes, _ := json.MarshalIndent(license, "", "  ")
	fmt.Println(string(jsonBytes))
}
