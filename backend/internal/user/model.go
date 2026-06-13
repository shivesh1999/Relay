package user

import "time"

type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    *string    `json:"password_hash,omitempty"`
	IsEmailVerified bool       `json:"is_email_verified"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

type UserProvider struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	ProviderData   *string   `json:"provider_data,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
