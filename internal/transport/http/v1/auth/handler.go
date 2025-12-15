package auth

import (
	"errors"
	"log/slog"
	domain "metrika/internal/domain/auth"
	"metrika/internal/infrastructure/jwt"
	"metrika/internal/usecase/auth"
	response "metrika/pkg/api"
	"metrika/pkg/logger/sl"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	log     *slog.Logger
	login   *auth.LoginUseCase
	refresh *auth.RefreshUseCase
	jwt     *jwt.JWTProvider
}

func NewHandler(log *slog.Logger, login *auth.LoginUseCase, refresh *auth.RefreshUseCase) *Handler {
	return &Handler{login: login, refresh: refresh, log: log}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	response.Response
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := render.Decode(r, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := response.ValidateRequest(req); err != nil {
		validateErr := err.(validator.ValidationErrors)
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.ValidationError(validateErr))
		return
	}

	tokens, err := h.login.Execute(
		r.Context(),
		req.Email,
		req.Password,
		r.UserAgent(),
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.jwt.SetCookie(w, tokens.Refresh)
	render.JSON(w, r, AuthResponse{Token: tokens.Access})
}

type RegisterRequest struct {
	Name           string `json:"name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	PasswordSecond string `json:"password_second" validate:"required"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := r.Cookie("refresh_token")
	if err != nil {
		h.log.Error("error", sl.Err(err))
	}
	if refresh_token == nil || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokens, err := h.refresh.Execute(
		r.Context(),
		refresh_token.Value,
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.jwt.SetCookie(w, tokens.Refresh)
	render.JSON(w, r, AuthResponse{Token: tokens.Access})
}
