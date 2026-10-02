-- name: GetActiveUserByLogin :one
SELECT id, login, password_hash, created_at, disabled_at
FROM users
WHERE login = $1
  AND disabled_at IS NULL;

-- name: GetActiveUserById :one
SELECT *
FROM users
WHERE id = $1
  AND disabled_at IS NULL;