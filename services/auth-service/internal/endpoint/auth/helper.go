package auth

import (
	"crypto/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const BcryptHashCost = 12

func generateDummyHash() (string, error) {
	password := make([]byte, 32)

	if _, err := rand.Read(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword(password, BcryptHashCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		BcryptHashCost,
	)
	return string(hash), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
	return err == nil
}

func SetCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Expires:  time.Now().UTC().Add(24 * time.Hour),
	})
}

func RemoveCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
}
