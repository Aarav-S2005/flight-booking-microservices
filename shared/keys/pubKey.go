package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func GetPublicKey(path string) (*ecdsa.PublicKey, error) {
	pubBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubBytes)
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey := parsed.(*ecdsa.PublicKey)
	return publicKey, nil
}
