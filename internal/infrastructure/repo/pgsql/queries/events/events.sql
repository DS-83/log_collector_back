-- name: CreateEvent :exec
INSERT INTO events (event_id, occurred_at, source, type, level, actor_id, request_id, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (occurred_at, event_id) DO NOTHING;

-- name: GetEventByID :one
SELECT id, event_id, occurred_at, received_at, source, type, level, actor_id, request_id, payload
FROM events
WHERE id = $1;

