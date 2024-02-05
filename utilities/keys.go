package utilities

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// Generates public & private PEM encoded keys for Folderr's usage in its authentication handling.
// Returns privateKey, publicKey, error
func GenKeys() ([]byte, []byte, error) {
	// generate keys
	// this is for private keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	if privateKey.Validate() != nil {
		return nil, nil, privateKey.Validate()
	}
	// Turned to a PKCS8 key so Folderr's JWT library (node-jwt) can hopefully read it
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privBytes,
	}

	privatePem := pem.EncodeToMemory(&privBlock)

	// We turn this key to a PKIX key so Folderr's JWT library (node-jwt) can hopefully read it
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	pubBlock := pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   pubBytes,
	}

	publicPem := pem.EncodeToMemory(&pubBlock)

	return privatePem, publicPem, nil
}
