package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
