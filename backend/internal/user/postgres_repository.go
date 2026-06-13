package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/relay/backend/internal/logger"
)

type PostgresRepository struct {
	db  *sql.DB
	log *logger.Logger
}

func NewPostgresRepository(db *sql.DB, log *logger.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:  db,
		log: log,
	}
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*User, error) {
	const query = `SELECT id, email, password_hash, is_email_verified, is_active, created_at, updated_at, last_login_at FROM users WHERE id = $1`

	return r.findOne(ctx, query, id, "query user by id")
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	const query = `SELECT id, email, password_hash, is_email_verified, is_active, created_at, updated_at, last_login_at FROM users WHERE email = $1`

	return r.findOne(ctx, query, email, "query user by email")
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg any, errMessage string) (*User, error) {
	row := r.db.QueryRowContext(ctx, query, arg)

	var u User
	var passwordHash sql.NullString

	if err := row.Scan(
		&u.ID,
		&u.Email,
		&passwordHash,
		&u.IsEmailVerified,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	if passwordHash.Valid {
		u.PasswordHash = &passwordHash.String
	}

	return &u, nil
}

func (r *PostgresRepository) Create(ctx context.Context, u *User) error {
	const query = `
INSERT INTO users (email, password_hash, is_email_verified, is_active, last_login_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at
`

	var passwordHash *string
	if u.PasswordHash != nil {
		passwordHash = u.PasswordHash
	}

	if err := r.db.QueryRowContext(ctx, query,
		u.Email,
		passwordHash,
		u.IsEmailVerified,
		u.IsActive,
		u.LastLoginAt,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *PostgresRepository) UpdateLastLoginAt(ctx context.Context, id string, loggedInAt time.Time) error {
	const query = `UPDATE users SET last_login_at = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, loggedInAt, id)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read last login update result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("update last login: user not found")
	}

	return nil
}
