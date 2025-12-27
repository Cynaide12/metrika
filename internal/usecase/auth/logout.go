package auth

import (
	"context"
	"log/slog"
	domain "metrika/internal/domain/auth"
	"metrika/internal/infrastructure/jwt"
)

type LogoutUseCase struct {
	sessions domain.SessionRepository
	log      *slog.Logger
	jwt jwt.JWTProvider
}

func NewLogoutUseCase(sessions domain.SessionRepository, log *slog.Logger, jwt jwt.JWTProvider) *LogoutUseCase {
	return &LogoutUseCase{
		sessions,
		log,
		jwt,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, refresh_token string) error {
	refreshClaims, err := uc.jwt.Validate(refresh_token)
	if err != nil {
		return domain.ErrInvalidRefreshToken
	}
	return uc.sessions.Delete(ctx, refreshClaims.SessionID)
}
