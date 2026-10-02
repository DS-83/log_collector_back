package pgsql

import (
	"context"
	"encoding/json"
	"errors"
	e "log_collect/err"
	"log_collect/internal/infrastructure/repo/pgsql/sqlcgen/eventsdb"
	"log_collect/models"
	"strings"

	"github.com/jackc/pgx/v5"
)

type EventRepo struct {
	query *eventsdb.Queries
	db    eventsdb.DBTX
}

func NewEventRepo(db eventsdb.DBTX) *EventRepo {
	return &EventRepo{
		query: eventsdb.New(db),
		db:    db,
	}
}

func (r *EventRepo) Create(ctx context.Context, e *models.Event) error {
	pb, err := payloadToByte(e.Payload)
	if err != nil {
		return err
	}

	arg := eventsdb.CreateEventParams{
		EventID:    e.EventID,
		OccurredAt: e.OccurredAt,
		Source:     e.Source,
		Type:       e.EventType,
		Level:      int16(e.Level),
		ActorID:    e.ActorID,
		RequestID:  e.RequestID,
		Payload:    pb,
	}

	return r.query.CreateEvent(ctx, arg)
}

func (r *EventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	row, err := r.query.GetEventByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, e.ErrNotFound
		}
		return nil, err
	}

	return eventToModel(row)
}

var allowedSortFields = map[string]string{
	"occurred_at": "occurred_at",
	"type":        "type",
	"source":      "source",
	"level":       "level",
}

const eventsWhere = `WHERE occurred_at >= $2 AND occurred_at < $3
	AND ($1::text = '' OR type ILIKE '%' || $1::text || '%' OR source ILIKE '%' || $1::text || '%')`

func (r *EventRepo) List(ctx context.Context, f *models.EventFilter) ([]models.Event, int, error) {
	col, ok := allowedSortFields[f.SortField]
	if !ok {
		col = "occurred_at"
	}
	dir := "ASC"
	if strings.EqualFold(f.Order, "desc") {
		dir = "DESC"
	}

	var total int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM events `+eventsWhere, f.Query, f.From, f.To).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	listSQL := `SELECT id, event_id, occurred_at, source, type, level FROM events ` + eventsWhere +
		` ORDER BY ` + col + ` ` + dir + `, id ` + dir + ` LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, listSQL, f.Query, f.From, f.To, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		var lvl int16
		if err := rows.Scan(&e.ID, &e.EventID, &e.OccurredAt, &e.Source, &e.EventType, &lvl); err != nil {
			return nil, 0, err
		}
		e.Level = models.EventLevel(lvl)
		events = append(events, e)
	}
	return events, total, rows.Err()
}

func payloadToByte(p map[string]any) ([]byte, error) {
	if p == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(p)
}

func payloadToMap(p []byte) (map[string]any, error) {
	if p == nil {
		return nil, nil
	}
	var data map[string]any

	err := json.Unmarshal(p, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func eventToModel(e eventsdb.Event) (*models.Event, error) {
	p, err := payloadToMap(e.Payload)
	if err != nil {
		return nil, err
	}

	return &models.Event{
		ID:         e.ID,
		EventID:    e.EventID,
		OccurredAt: e.OccurredAt,
		ReceivedAt: e.ReceivedAt,
		Source:     e.Source,
		EventType:  e.Type,
		Level:      models.EventLevel(e.Level),
		ActorID:    e.ActorID,
		RequestID:  e.RequestID,
		Payload:    p,
	}, nil
}
