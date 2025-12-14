package handlers

import (
	"metrika/internal/models"
	response "metrika/lib/api"
	"net/http"
)

type AuthService interface {
	Register(user *models.User, userAgent string) (accessToken string, refreshToken string, err error)
	Login(email string, password string, userAgent string) (accessToken string, refreshToken string, err error)
	RefreshTokens(refreshTokenCookie *http.Cookie) (accessToken string, refreshToken string, err error)
	Logout(session_id uint) (err error)
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name           string `json:"name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	PasswordSecond string `json:"password_second" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	response.Response
}
