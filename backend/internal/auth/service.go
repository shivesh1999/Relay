package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/relay/backend/internal/config"
	"github.com/relay/backend/internal/logger"
	"github.com/relay/backend/internal/user"
)

type Service struct {
	userRepo       user.Repository
	log            *logger.Logger
	jwtSecret      []byte
	accessTokenTTL time.Duration
}

type accessTokenClaims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	TokenType string `json:"typ"`
}

func NewService(userRepo user.Repository, log *logger.Logger, authCfg config.AuthConfig) *Service {
	return &Service{
		userRepo:       userRepo,
		log:            log,
		jwtSecret:      []byte(authCfg.JWTSecret),
		accessTokenTTL: time.Duration(authCfg.AccessTokenTTLMinutes) * time.Minute,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	email := strings.TrimSpace(req.Email)
	password := req.Password

	if email == "" {
		return RegisterResponse{}, ValidationError{Field: "email", Message: "email is required"}
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return RegisterResponse{}, ValidationError{Field: "email", Message: "invalid email format"}
	}

	if password == "" {
		return RegisterResponse{}, ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < 8 {
		return RegisterResponse{}, ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return RegisterResponse{}, err
	}

	if existing != nil {
		return RegisterResponse{}, EmailAlreadyExistsError{Email: email}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, err
	}

	passwordHash := string(hashed)
	now := time.Now().UTC()

	newUser := &user.User{
		Email:           email,
		PasswordHash:    &passwordHash,
		IsEmailVerified: false,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		ID:        newUser.ID,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	email := strings.TrimSpace(req.Email)
	password := req.Password

	if email == "" {
		return LoginResponse{}, ValidationError{Field: "email", Message: "email is required"}
	}

	if password == "" {
		return LoginResponse{}, ValidationError{Field: "password", Message: "password is required"}
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return LoginResponse{}, err
	}

	if existing == nil || existing.PasswordHash == nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	if !existing.IsActive {
		return LoginResponse{}, ErrInactiveUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*existing.PasswordHash), []byte(password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenTTL)
	token, err := s.generateAccessToken(existing, now, expiresAt)
	if err != nil {
		return LoginResponse{}, err
	}

	if err := s.userRepo.UpdateLastLoginAt(ctx, existing.ID, now); err != nil {
		return LoginResponse{}, err
	}
	existing.LastLoginAt = &now

	return LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        toUserResponse(existing),
	}, nil
}

func (s *Service) AuthenticateToken(ctx context.Context, token string) (*user.User, error) {
	claims, err := s.parseAccessToken(token)
	if err != nil {
		return nil, err
	}

	existing, err := s.userRepo.FindByID(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, ErrInvalidToken
	}

	if !existing.IsActive {
		return nil, ErrInactiveUser
	}

	return existing, nil
}

func (s *Service) generateAccessToken(u *user.User, issuedAt time.Time, expiresAt time.Time) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := accessTokenClaims{
		Subject:   u.ID,
		Email:     u.Email,
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: expiresAt.Unix(),
		TokenType: "access",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	unsignedToken := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	signature := s.sign(unsignedToken)

	return unsignedToken + "." + signature, nil
}

func (s *Service) parseAccessToken(token string) (accessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return accessTokenClaims{}, ErrInvalidToken
	}

	unsignedToken := parts[0] + "." + parts[1]
	expectedSignature := s.sign(unsignedToken)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return accessTokenClaims{}, ErrInvalidToken
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return accessTokenClaims{}, ErrInvalidToken
	}

	var claims accessTokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return accessTokenClaims{}, ErrInvalidToken
	}

	if claims.Subject == "" || claims.TokenType != "access" {
		return accessTokenClaims{}, ErrInvalidToken
	}

	if time.Now().UTC().Unix() >= claims.ExpiresAt {
		return accessTokenClaims{}, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) sign(unsignedToken string) string {
	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write([]byte(unsignedToken))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func toUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		IsEmailVerified: u.IsEmailVerified,
		IsActive:        u.IsActive,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		LastLoginAt:     u.LastLoginAt,
	}
}

func IsAuthError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrInactiveUser) ||
		errors.Is(err, ErrInvalidToken)
}
