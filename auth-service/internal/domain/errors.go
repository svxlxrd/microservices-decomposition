package domain

import "errors"

var (
	// Auth errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameExists     = errors.New("username already exists")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrUserNotFound       = errors.New("user not found")
)