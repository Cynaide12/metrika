package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"metrika/internal/config"
	lib_jwt "metrika/lib/jwt"
	"strings"

	"github.com/go-chi/render"
)

var (
	JWTClaimsDataKey = "jwt_claims_key"
	TokenEmpty       = "empty token"
	InvalidToken     = "invalid token"
)

// AuthMiddleware - мидлвэйр с авторизацией
func AuthMiddleware(log *slog.Logger, jwtSecret string, cfg config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log = log.With(
				slog.String("component", "middleware/authMiddleware"),
			)

			accessToken := r.Header.Get("Authorization")
			if accessToken == "" {
				log.Error("access token is empty")
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, TokenEmpty)
				return
			}

			authParts := strings.Split(accessToken, " ")
			if len(authParts) < 2 {
				log.Error("access token is empty")
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, InvalidToken)
				return
			}

			claims, err := lib_jwt.ValidateJWT(cfg.JWTSecret, authParts[1])
			if err != nil {
				//если access токен невалид - отправляем 401 статус чтобы фронт послал запрос на обновление аксес токена по рефреш токену
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, "you need to update access token")
				return
			}

			// если access токен валиден, продолжаем цепочку запроса
			c := context.WithValue(r.Context(), JWTClaimsDataKey, claims)

			//чтобы в логи данные юзера записывались - email и user_id
			log.Info("claims:", slog.Any("claims", claims))

			next.ServeHTTP(w, r.WithContext(c))
		})
	}
}
