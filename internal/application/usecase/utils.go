package usecase

import (
	"crypto/rand"
	"encoding/base64"
)

func generateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil { // crypto/rand
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
