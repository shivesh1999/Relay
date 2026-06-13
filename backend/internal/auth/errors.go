package auth

import (
	"errors"
	"fmt"
)

var ErrEmailAlreadyExists = errors.New("email already registered")
var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrInactiveUser = errors.New("user account is inactive")
var ErrInvalidToken = errors.New("invalid or expired token")

type EmailAlreadyExistsError struct {
	Email string
}

func (e EmailAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s: %s", ErrEmailAlreadyExists.Error(), e.Email)
}

func (e EmailAlreadyExistsError) Is(target error) bool {
	return target == ErrEmailAlreadyExists
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
