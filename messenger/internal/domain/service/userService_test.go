package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"messanger/internal/domain/entity"
)

type fakeUserRepo struct {
	users map[string]*entity.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*entity.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, user *entity.User) error {
	f.users[user.Username] = user
	return nil
}

func (f *fakeUserRepo) GetByUUID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	for _, u := range f.users {
		if u.UUID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) GetByUsername(_ context.Context, username string) (*entity.User, error) {
	if u, ok := f.users[username]; ok {
		return u, nil
	}
	return nil, nil
}

func (f *fakeUserRepo) Update(_ context.Context, user *entity.User) error { return nil }

func (f *fakeUserRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeUserRepo) GetAll(_ context.Context) ([]*entity.User, error) {
	var out []*entity.User
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

func TestCreateUserHashesPassword(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())

	user, err := svc.CreateUser(context.Background(), "alice", "secret-password")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if user.PasswordHash == "secret-password" {
		t.Fatal("password stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret-password")); err != nil {
		t.Fatalf("hash does not match password: %v", err)
	}
}

func TestCreateUserShortPassword(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())

	if _, err := svc.CreateUser(context.Background(), "alice", "short"); err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestAuthenticateSuccess(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())
	if _, err := svc.CreateUser(context.Background(), "alice", "secret-password"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	user, err := svc.Authenticate(context.Background(), "alice", "secret-password")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("unexpected user: %s", user.Username)
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())
	if _, err := svc.CreateUser(context.Background(), "alice", "secret-password"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if _, err := svc.Authenticate(context.Background(), "alice", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticateUnknownUser(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())

	if _, err := svc.Authenticate(context.Background(), "ghost", "whatever-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticateLegacyUserWithoutPassword(t *testing.T) {
	repo := newFakeUserRepo()
	// Пользователь, созданный до миграции 004: пустой хэш
	repo.users["legacy"] = &entity.User{UUID: uuid.New(), Username: "legacy"}
	svc := NewUserService(repo)

	if _, err := svc.Authenticate(context.Background(), "legacy", ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials for empty hash, got %v", err)
	}
}
