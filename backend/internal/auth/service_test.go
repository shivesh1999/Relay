package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/relay/backend/internal/config"
	"github.com/relay/backend/internal/user"
)

type fakeUserRepository struct {
	byID    map[string]*user.User
	byEmail map[string]*user.User
}

func newFakeUserRepository(users ...*user.User) *fakeUserRepository {
	repo := &fakeUserRepository{
		byID:    make(map[string]*user.User),
		byEmail: make(map[string]*user.User),
	}

	for _, u := range users {
		repo.byID[u.ID] = u
		repo.byEmail[u.Email] = u
	}

	return repo
}

func (r *fakeUserRepository) FindByID(_ context.Context, id string) (*user.User, error) {
	return r.byID[id], nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	return r.byEmail[email], nil
}

func (r *fakeUserRepository) Create(_ context.Context, u *user.User) error {
	if u.ID == "" {
		u.ID = "user-created"
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *fakeUserRepository) UpdateLastLoginAt(_ context.Context, id string, loggedInAt time.Time) error {
	u := r.byID[id]
	if u == nil {
		return errors.New("user not found")
	}
	u.LastLoginAt = &loggedInAt
	return nil
}

func TestLoginIssuesTokenAndAuthenticatesUser(t *testing.T) {
	passwordHash := hashPassword(t, "password123")
	existingUser := &user.User{
		ID:              "user-1",
		Email:           "shivesh@example.com",
		PasswordHash:    &passwordHash,
		IsEmailVerified: true,
		IsActive:        true,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	service := NewService(newFakeUserRepository(existingUser), nil, config.AuthConfig{
		JWTSecret:             "test-secret",
		AccessTokenTTLMinutes: 15,
	})

	resp, err := service.Login(context.Background(), LoginRequest{
		Email:    "shivesh@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("expected access token")
	}

	if resp.TokenType != "Bearer" {
		t.Fatalf("expected bearer token type, got %q", resp.TokenType)
	}

	if resp.User.ID != existingUser.ID {
		t.Fatalf("expected user id %q, got %q", existingUser.ID, resp.User.ID)
	}

	if existingUser.LastLoginAt == nil {
		t.Fatal("expected last login timestamp to be updated")
	}

	currentUser, err := service.AuthenticateToken(context.Background(), resp.AccessToken)
	if err != nil {
		t.Fatalf("authenticate token returned error: %v", err)
	}

	if currentUser.ID != existingUser.ID {
		t.Fatalf("expected authenticated user id %q, got %q", existingUser.ID, currentUser.ID)
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	passwordHash := hashPassword(t, "password123")
	existingUser := &user.User{
		ID:           "user-1",
		Email:        "shivesh@example.com",
		PasswordHash: &passwordHash,
		IsActive:     true,
	}

	service := NewService(newFakeUserRepository(existingUser), nil, config.AuthConfig{
		JWTSecret:             "test-secret",
		AccessTokenTTLMinutes: 15,
	})

	_, err := service.Login(context.Background(), LoginRequest{
		Email:    "shivesh@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestAuthenticateTokenRejectsExpiredToken(t *testing.T) {
	existingUser := &user.User{
		ID:       "user-1",
		Email:    "shivesh@example.com",
		IsActive: true,
	}

	service := NewService(newFakeUserRepository(existingUser), nil, config.AuthConfig{
		JWTSecret:             "test-secret",
		AccessTokenTTLMinutes: 15,
	})

	token, err := service.generateAccessToken(
		existingUser,
		time.Now().UTC().Add(-2*time.Hour),
		time.Now().UTC().Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf("generate access token returned error: %v", err)
	}

	_, err = service.AuthenticateToken(context.Background(), token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	return string(hash)
}
