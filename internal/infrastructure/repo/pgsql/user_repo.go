package pgsql

import (
	"context"
	"errors"
	e "log_collect/err"
	"log_collect/internal/infrastructure/repo/pgsql/sqlcgen/usersdb"
	"log_collect/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	query *usersdb.Queries
}

func NewUserRepo(db usersdb.DBTX) *UserRepo {
	return &UserRepo{query: usersdb.New(db)}
}

func (u *UserRepo) GetActiveByLogin(ctx context.Context, l string) (*models.User, error) {
	row, err := u.query.GetActiveUserByLogin(ctx, l)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, e.ErrNotFound
		}
		return nil, err
	}
	return userToModel(row), nil
}

func (u *UserRepo) GetActiveByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row, err := u.query.GetActiveUserById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, e.ErrNotFound
		}
		return nil, err
	}
	return userToModel(row), nil
}

func userToModel(r usersdb.User) *models.User {
	return &models.User{
		ID:           r.ID,
		Login:        r.Login,
		PasswordHash: r.PasswordHash,
		CreatedAt:    r.CreatedAt,
		DisabledAt:   r.DisabledAt,
	}
}
