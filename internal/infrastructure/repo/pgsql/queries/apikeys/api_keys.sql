-- name: GetActiveApiKeyByHash :one
SELECT id, name, key_hash, source, created_at, revoked_at
FROM api_keys
WHERE key_hash = $1
  AND revoked_at IS NULL;

-- name: CreateApiKey :exec
INSERT INTO api_keys (name, key_hash, source)
VALUES ($1, $2, $3);