package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	TokenHash []byte // sha256 токена из cookie
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
}
