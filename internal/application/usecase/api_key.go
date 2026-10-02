package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log_collect/internal"
	"log_collect/models"
)

type ApiKeyUseCase struct {
	repo internal.ApiKeyRepo
}

func NewApiKeyUseCase(r internal.ApiKeyRepo) *ApiKeyUseCase {
	return &ApiKeyUseCase{
		repo: r,
	}
}

func (u *ApiKeyUseCase) Create(ctx context.Context, name, source string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	rawKey := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(rawKey))

	k := &models.ApiKey{
		Name:    name,
		KeyHash: hash[:],
		Source:  source,
	}
	if err := u.repo.Create(ctx, k); err != nil {
		return "", err
	}

	return rawKey, nil
}

func (u *ApiKeyUseCase) Authenticate(ctx context.Context, rawKey string) (*models.ApiKey, error) {
	hash := sha256.Sum256([]byte(rawKey))
	return u.repo.GetActiveByHash(ctx, hash[:])
}
