package internal

import (
	"context"
	"log_collect/models"
)

type ApiKeyUseCase interface {
	Create(ctx context.Context, name, source string) (string, error)
	Authenticate(ctx context.Context, rawKey string) (*models.ApiKey, error)
}

type UserUseCase interface {
	Login(ctx context.Context, login, pass string) (token string, err error)
	Logout(ctx context.Context, token string) error
	Authenticate(ctx context.Context, token string) (*models.User, error)
}

type EventUseCase interface {
	Create(ctx context.Context, e *models.Event) error
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	List(ctx context.Context, f *models.EventFilter) ([]models.Event, int, error)
}
