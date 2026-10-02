package usecase

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	e "log_collect/err"
	"log_collect/models"

	"github.com/google/uuid"
)

// --- mocks ---

type mockUserRepo struct {
	getActiveByLogin func(ctx context.Context, login string) (*models.User, error)
	getActiveByID    func(ctx context.Context, id uuid.UUID) (*models.User, error)
}

func (m *mockUserRepo) GetActiveByLogin(ctx context.Context, login string) (*models.User, error) {
	return m.getActiveByLogin(ctx, login)
}

func (m *mockUserRepo) GetActiveByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.getActiveByID(ctx, id)
}

type mockSessionRepo struct {
	created           *models.Session // захватываем, что реально передали в Create
	createErr         error
	deleteByTokenHash func(ctx context.Context, hash []byte) error
	getActiveByToken  func(ctx context.Context, hash []byte) (*models.Session, error)
}

func (m *mockSessionRepo) Create(ctx context.Context, s *models.Session) error {
	m.created = s
	return m.createErr
}

func (m *mockSessionRepo) DeleteByTokenHash(ctx context.Context, hash []byte) error {
	return m.deleteByTokenHash(ctx, hash)
}

func (m *mockSessionRepo) GetActiveByTokenHash(ctx context.Context, hash []byte) (*models.Session, error) {
	return m.getActiveByToken(ctx, hash)
}

type mockHasher struct {
	verify func(password, hash string) (bool, error)
	hash   func(password string) (string, error)
}

func (m *mockHasher) Verify(password, hash string) (bool, error) {
	return m.verify(password, hash)
}

func (m *mockHasher) Hash(password string) (string, error) {
	return m.hash(password)
}

func TestAuthUseCase_Login(t *testing.T) {
	user := &models.User{
		ID:           uuid.New(),
		Login:        "alice",
		PasswordHash: "stored-hash",
	}
	dbErr := errors.New("db is down")

	tests := []struct {
		name string

		login    string
		password string

		getUser   func(ctx context.Context, login string) (*models.User, error)
		verify    func(password, hash string) (bool, error)
		createErr error

		wantErr       error
		wantToken     bool // ожидаем непустой токен
		wantSession   bool // ожидаем вызов sessionRepo.Create
		wantHasherHit bool // ожидаем вызов hasher.Verify
	}{
		{
			name:     "success",
			login:    "alice",
			password: "correct-password",
			getUser: func(ctx context.Context, login string) (*models.User, error) {
				return user, nil
			},
			verify: func(password, hash string) (bool, error) {
				return password == "correct-password" && hash == user.PasswordHash, nil
			},
			wantErr:       nil,
			wantToken:     true,
			wantSession:   true,
			wantHasherHit: true,
		},
		{
			name:     "user not found",
			login:    "ghost",
			password: "whatever",
			getUser: func(ctx context.Context, login string) (*models.User, error) {
				return nil, e.ErrNotFound
			},
			verify: func(password, hash string) (bool, error) {
				return false, nil
			},
			wantErr:       e.ErrInvalidCredentials,
			wantToken:     false,
			wantSession:   false,
			wantHasherHit: false, // не должны звать hasher, если юзера нет
		},
		{
			name:     "wrong password",
			login:    "alice",
			password: "wrong-password",
			getUser: func(ctx context.Context, login string) (*models.User, error) {
				return user, nil
			},
			verify: func(password, hash string) (bool, error) {
				return false, nil
			},
			wantErr:       e.ErrInvalidCredentials,
			wantToken:     false,
			wantSession:   false,
			wantHasherHit: true,
		},
		{
			name:     "hasher error treated as invalid credentials",
			login:    "alice",
			password: "any-password",
			getUser: func(ctx context.Context, login string) (*models.User, error) {
				return user, nil
			},
			verify: func(password, hash string) (bool, error) {
				return false, errors.New("corrupted hash format")
			},
			wantErr:       e.ErrInvalidCredentials,
			wantToken:     false,
			wantSession:   false,
			wantHasherHit: true,
		},
		{
			name:     "session create fails propagates raw error",
			login:    "alice",
			password: "correct-password",
			getUser: func(ctx context.Context, login string) (*models.User, error) {
				return user, nil
			},
			verify: func(password, hash string) (bool, error) {
				return true, nil
			},
			createErr:     dbErr,
			wantErr:       dbErr,
			wantToken:     false,
			wantSession:   true, // Create вызван, просто вернул ошибку
			wantHasherHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasherHit := false
			userRepo := &mockUserRepo{getActiveByLogin: tt.getUser}
			hasher := &mockHasher{
				verify: func(password, hash string) (bool, error) {
					hasherHit = true
					return tt.verify(password, hash)
				},
			}
			sessionRepo := &mockSessionRepo{createErr: tt.createErr}

			uc := NewAuthUseCase(userRepo, sessionRepo, hasher, 10*time.Minute)

			token, err := uc.Login(context.Background(), tt.login, tt.password)

			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
			if tt.wantToken && token == "" {
				t.Fatal("expected non-empty token")
			}
			if !tt.wantToken && token != "" {
				t.Fatal("expected empty token")
			}
			if tt.wantSession && sessionRepo.created == nil {
				t.Fatal("expected sessionRepo.Create to be called")
			}
			if !tt.wantSession && sessionRepo.created != nil {
				t.Fatal("expected sessionRepo.Create NOT to be called")
			}
			if tt.wantHasherHit != hasherHit {
				t.Fatalf("hasher called = %v, want %v", hasherHit, tt.wantHasherHit)
			}

			// доп. проверка инварианта только для успешного случая
			if tt.name == "success" {
				wantHash := sha256.Sum256([]byte(token))
				if string(sessionRepo.created.TokenHash) != string(wantHash[:]) {
					t.Fatal("stored TokenHash does not match sha256(token)")
				}
				if sessionRepo.created.UserID != user.ID {
					t.Fatalf("session created with wrong UserID: %v", sessionRepo.created.UserID)
				}
			}
		})
	}
}
