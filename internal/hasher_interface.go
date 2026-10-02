package internal

type Hasher interface {
	Hash(password string) (hash string, err error)
	Verify(password, hash string) (bool, error)
}
