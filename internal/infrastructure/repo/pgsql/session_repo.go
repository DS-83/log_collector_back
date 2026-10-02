package pgsql

import (
	"context"
	"errors"
	e "log_collect/err"
	"log_collect/internal/infrastructure/repo/pgsql/sqlcgen/sessionsdb"
	"log_collect/models"

	"github.com/jackc/pgx/v5"
)

type SessionRepo struct {
	query *sessionsdb.Queries
}

func NewSessionRepo(db sessionsdb.DBTX) *SessionRepo {
	return &SessionRepo{
		query: sessionsdb.New(db),
	}
}

func (r *SessionRepo) Create(ctx context.Context, s *models.Session) error {
	arg := sessionsdb.CreateSessionParams{
		TokenHash: s.TokenHash,
		UserID:    s.UserID,
		ExpiresAt: s.ExpiresAt,
	}

	return r.query.CreateSession(ctx, arg)
}

func (r *SessionRepo) DeleteByTokenHash(ctx context.Context, h []byte) error {
	return r.query.DeleteSessionByHash(ctx, h)
}

func (r *SessionRepo) GetActiveByTokenHash(ctx context.Context, h []byte) (*models.Session, error) {
	row, err := r.query.GetActiveSessionByHash(ctx, h)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, e.ErrNotFound
		}
		return nil, err
	}

	return sessionToModel(row), nil
}

func sessionToModel(s sessionsdb.Session) *models.Session {
	return &models.Session{
		TokenHash: s.TokenHash,
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt,
		ExpiresAt: s.ExpiresAt,
	}
}
