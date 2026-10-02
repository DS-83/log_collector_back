package internal

import (
	"context"
	"log_collect/models"

	"github.com/google/uuid"
)

type ApiKeyRepo interface {
	GetActiveByHash(ctx context.Context, hash []byte) (*models.ApiKey, error)
	Create(ctx context.Context, k *models.ApiKey) error
}

type UserRepo interface {
	GetActiveByLogin(ctx context.Context, login string) (*models.User, error)
	GetActiveByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type SessionRepo interface {
	Create(ctx context.Context, s *models.Session) error
	DeleteByTokenHash(ctx context.Context, h []byte) error
	GetActiveByTokenHash(ctx context.Context, h []byte) (*models.Session, error)
}

type EventRepo interface {
	Create(ctx context.Context, e *models.Event) error
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	List(ctx context.Context, f *models.EventFilter) ([]models.Event, int, error)
}
