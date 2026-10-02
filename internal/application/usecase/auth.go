package usecase

import (
	"context"
	"crypto/sha256"
	e "log_collect/err"
	"log_collect/internal"
	"log_collect/models"
	"time"
)

type AuthUseCase struct {
	userRepo    internal.UserRepo
	sessionRepo internal.SessionRepo
	hasher      internal.Hasher
	sessionTTL  time.Duration
}

func NewAuthUseCase(ur internal.UserRepo, sr internal.SessionRepo, h internal.Hasher, ttl time.Duration) *AuthUseCase {
	return &AuthUseCase{
		userRepo:    ur,
		sessionRepo: sr,
		hasher:      h,
		sessionTTL:  ttl,
	}
}

func (u *AuthUseCase) Login(ctx context.Context, login, pass string) (string, error) {
	user, err := u.userRepo.GetActiveByLogin(ctx, login)
	if err != nil {
		return "", e.ErrInvalidCredentials
	}
	ok, err := u.hasher.Verify(pass, user.PasswordHash)
	if err != nil || !ok {
		return "", e.ErrInvalidCredentials
	}

	rawToken, err := generateToken()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(rawToken))
	expiresAt := time.Now().Add(u.sessionTTL)

	err = u.sessionRepo.Create(ctx, &models.Session{
		TokenHash: hash[:],
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

func (u *AuthUseCase) Logout(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	return u.sessionRepo.DeleteByTokenHash(ctx, hash[:])
}

func (u *AuthUseCase) Authenticate(ctx context.Context, token string) (*models.User, error) {
	hash := sha256.Sum256([]byte(token))
	session, err := u.sessionRepo.GetActiveByTokenHash(ctx, hash[:])
	if err != nil {
		return nil, e.ErrUnauthenticated
	}

	return u.userRepo.GetActiveByID(ctx, session.UserID)
}
