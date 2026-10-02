package pgsql

import (
	"context"
	"errors"
	e "log_collect/err"
	"log_collect/internal/infrastructure/repo/pgsql/sqlcgen/apikeysdb"
	"log_collect/models"

	"github.com/jackc/pgx/v5"
)

type ApiKeyRepo struct {
	query *apikeysdb.Queries
}

// sqlcgen.DBTX удовлетворяют и *pgxpool.Pool, и pgx.Tx
func NewApiKeyRepo(db apikeysdb.DBTX) *ApiKeyRepo {
	return &ApiKeyRepo{query: apikeysdb.New(db)}
}

func (r *ApiKeyRepo) GetActiveByHash(ctx context.Context, hash []byte) (*models.ApiKey, error) {
	row, err := r.query.GetActiveApiKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, e.ErrNotFound
		}
		return nil, err
	}
	key := apiKeyToModel(row)
	return key, nil
}

func (r *ApiKeyRepo) Create(ctx context.Context, k *models.ApiKey) error {
	return r.query.CreateApiKey(ctx, apikeysdb.CreateApiKeyParams{
		Name:    k.Name,
		KeyHash: k.KeyHash,
		Source:  k.Source,
	})
}

func apiKeyToModel(r apikeysdb.ApiKey) *models.ApiKey {
	return &models.ApiKey{
		ID:        r.ID,
		Name:      r.Name,
		KeyHash:   r.KeyHash,
		Source:    r.Source,
		CreatedAt: r.CreatedAt,
		RevokedAt: r.RevokedAt,
	}
}
