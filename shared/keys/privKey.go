package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func GetPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	privBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privBytes)
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, err
}
