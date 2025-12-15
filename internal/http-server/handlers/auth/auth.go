package handlers

import (
	"errors"
	"log/slog"
	middleware "metrika/internal/http-server/middlewares"
	"metrika/internal/models"
	"metrika/internal/service"
	response "metrika/lib/api"
	lib_cookie "metrika/lib/cookie"
	lib_jwt "metrika/lib/jwt"
	"metrika/lib/logger/sl"
	lib_password "metrika/lib/password"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type authhandler struct {
	log     *slog.Logger
	service AuthService
}

func NewAuthHandler(log *slog.Logger, service AuthService, r chi.Router) *authhandler {

	h := &authhandler{
		log:     log,
		service: service,
	}

	r.Get("/api/v1/auth/login", h.Login())
	r.Get("/api/v1/auth/register", h.Register())
	r.Get("/api/v1/auth/refresh", h.Refresh())
	r.Get("/api/v1/auth/logout", h.Logout())

	return h
}

// func (h *authhandler) Login() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		const fn = "http-server.handlers.auth.Login"
// 		logger := h.log.With(slog.String("fn", fn))

// 		var req LoginRequest

// 		// декодируем запрос в структуру
// 		if err := render.Decode(r, &req); err != nil {
// 			logger.Error("failed to decode request", sl.Err(err))
// 			w.WriteHeader(http.StatusBadRequest)
// 			render.JSON(w, r, response.Error("invalid request"))
// 			return
// 		}

// 		// валидация запроса
// 		if err := response.ValidateRequest(&req); err != nil {
// 			validateErr := err.(validator.ValidationErrors)
// 			logger.Error("invalid request", sl.Err(err))
// 			w.WriteHeader(http.StatusBadRequest)
// 			render.JSON(w, r, response.ValidationError(validateErr))
// 			return
// 		}

// 		accessToken, refreshToken, err := h.service.Login(req.Email, req.Password, r.UserAgent())
// 		if err != nil {
// 			if errors.Is(err, service.ErrAlreadyExists) {
// 				w.WriteHeader(http.StatusUnauthorized)
// 				render.JSON(w, r, "user not found")
// 				return
// 			}
// 			if errors.Is(err, service.Unauthorized) {
// 				w.WriteHeader(http.StatusUnauthorized)
// 				render.JSON(w, r, "unathorized")
// 				return
// 			}
// 			logger.Error("failed to login", sl.Err(err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			render.JSON(w, r, response.Error("failed to login"))
// 			return
// 		}

// 		//установка рефреш токена в куки
// 		lib_cookie.AddCookie(w, refreshToken)

// 		render.JSON(w, r, AuthResponse{
// 			Token:    accessToken,
// 			Response: response.OK(),
// 		})

// 	}
// }

func (h *authhandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.auth.Register"
		logger := h.log.With(slog.String("fn", fn))

		var req RegisterRequest

		// декодируем запрос в структуру
		if err := render.Decode(r, &req); err != nil {
			logger.Error("failed to decode request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		// валидация запроса
		if err := response.ValidateRequest(&req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			logger.Error("invalid request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		if req.Password != req.PasswordSecond {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("password dont match"))
			return
		}

		hashPassword, err := lib_password.HashPassword(req.PasswordSecond)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to parse password"))
			return
		}

		user := models.User{
			Name:     req.Name,
			Email:    req.Email,
			Password: hashPassword,
		}

		accessToken, refreshToken, err := h.service.Register(&user, r.UserAgent())
		if err != nil {
			if errors.Is(err, service.UserAlreadyExsits) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, "user with this email already exists")
				return
			}
			logger.Error("failed to register user", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to register"))
			return
		}

		//установка рефреш токена в куки
		lib_cookie.AddCookie(w, refreshToken)

		render.JSON(w, r, AuthResponse{
			Token:    accessToken,
			Response: response.OK(),
		})

	}
}

// // получение нового access токена по refresh токену и замена refresh токена
// func (h *authhandler) Refresh() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		const fn = "http-server.handlers.auth.Refresh"
// 		logger := h.log.With(slog.String("fn", fn))

// 		refresh_token, err := r.Cookie("refresh_token")
// 		if err != nil {
// 			logger.Error("error", sl.Err(err))
// 		}
// 		if refresh_token == nil || err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		accessToken, refreshToken, err := h.service.RefreshTokens(refresh_token)
// 		if err != nil {
// 			if errors.Is(err, service.UserNotFound) {
// 				w.WriteHeader(http.StatusBadRequest)
// 				render.JSON(w, r, "user not found")
// 				return
// 			}
// 			if errors.Is(err, service.Unauthorized) || errors.Is(err, service.RefreshTokenInvalid) {
// 				w.WriteHeader(http.StatusUnauthorized)
// 				render.JSON(w, r, "unathorized")
// 				return
// 			}
// 			logger.Error("failed to refresh tokens", sl.Err(err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			render.JSON(w, r, response.Error("failed to refresh tokens"))
// 			return
// 		}

// 		//установка нового рефреш токена в куки
// 		lib_cookie.AddCookie(w, refreshToken)

// 		render.JSON(w, r, AuthResponse{
// 			Token:    accessToken,
// 			Response: response.OK(),
// 		})

// 	}
// }

func (h *authhandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.auth.Logout"
		logger := h.log.With(slog.String("fn", fn))

		refresh_token, err := r.Cookie("refresh_token")
		if err != nil {
			logger.Error("error", sl.Err(err))
		}
		if refresh_token.Value == "" || err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		//удаляем рефреш токен из куков
		lib_cookie.RemoveCookie(w)

		userClaims, ok := r.Context().Value(middleware.JWTClaimsDataKey).(*lib_jwt.JWTClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := h.service.Logout(userClaims.SessionID); err != nil {
			if errors.Is(err, service.UserSessionNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, service.Unauthorized)
				return
			}
			logger.Error("failed to logout", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to logout"))
			return
		}

		render.JSON(w, r, response.OK())

	}
}
