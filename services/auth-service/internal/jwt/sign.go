package jwt

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var tokenAuth *jwtauth.JWTAuth

func InitAuth(privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey) {
	tokenAuth = jwtauth.New("ES256", privKey, pubKey)
}

func SignJwt(id uuid.UUID) (string, error) {
	if tokenAuth == nil {
		return "", errors.New("auth not initialized")
	}
	now := time.Now()

	_, token, err := tokenAuth.Encode(map[string]interface{}{
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 1).Unix(),
		"iss": "auth-service",
		"sub": id.String(),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}
