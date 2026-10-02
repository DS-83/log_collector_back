-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: DeleteSessionByHash :exec
DELETE FROM sessions
WHERE token_hash = $1;

-- name: GetActiveSessionByHash :one
SELECT token_hash, user_id, created_at, expires_at
FROM sessions
WHERE token_hash = $1
  AND expires_at > now();

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;