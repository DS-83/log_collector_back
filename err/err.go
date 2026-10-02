package e

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidReqData     = errors.New("invalid request data")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("unauthenticated")

	ErrDbDsnNotSet = errors.New("DB_DSN is not set")
)
