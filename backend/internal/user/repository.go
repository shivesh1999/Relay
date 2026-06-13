package user

import (
	"context"
	"time"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
	UpdateLastLoginAt(ctx context.Context, id string, loggedInAt time.Time) error
}
