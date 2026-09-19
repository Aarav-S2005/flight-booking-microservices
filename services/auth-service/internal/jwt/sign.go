package jwt

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

func InitAuth(privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey) *jwtauth.JWTAuth {
	return jwtauth.New("ES256", privKey, pubKey)
}

func SignJwt(tokenAuth *jwtauth.JWTAuth, id uuid.UUID) (string, error) {
	if tokenAuth == nil {
		return "", errors.New("auth not initialized")
	}
	now := time.Now().UTC()

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
