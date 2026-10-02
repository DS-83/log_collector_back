package models

import (
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	ID        uuid.UUID
	Name      string
	KeyHash   []byte // sha256 ключа, сам ключ не хранится
	Source    string
	CreatedAt time.Time
	RevokedAt *time.Time
}
