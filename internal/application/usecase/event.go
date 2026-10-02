package usecase

import (
	"context"
	e "log_collect/err"
	"log_collect/internal"
	"log_collect/models"
	"time"
)

type EventUseCase struct {
	repo internal.EventRepo
}

func NewEventUseCase(r internal.EventRepo) *EventUseCase {
	return &EventUseCase{
		repo: r,
	}
}

func (r *EventUseCase) Create(ctx context.Context, e *models.Event) error {
	return r.repo.Create(ctx, e)
}

func (r *EventUseCase) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	return r.repo.GetByID(ctx, id)
}

const (
	defaultRange = 7 * 24 * time.Hour
	maxRange     = 31 * 24 * time.Hour
)

func (u *EventUseCase) List(ctx context.Context, in *models.EventFilter) ([]models.Event, int, error) {
	f := *in
	if f.To.IsZero() {
		f.To = time.Now()
	}
	if f.From.IsZero() {
		f.From = f.To.Add(-defaultRange)
	}
	if !f.From.Before(f.To) || f.To.Sub(f.From) > maxRange {
		return nil, 0, e.ErrInvalidReqData
	}
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		return nil, 0, e.ErrInvalidReqData
	}
	return u.repo.List(ctx, &f)
}
