package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidPasswordRaw  = errors.New("invalid password raw")
	ErrPasswordNotCompare  = errors.New("err password not compare")
)
