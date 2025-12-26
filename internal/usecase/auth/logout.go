package auth

import (
	"context"
	"log/slog"
	domain "metrika/internal/domain/auth"
)

type LogoutUseCase struct {
	sessions domain.SessionRepository
	log      *slog.Logger
}

func NewLogoutUseCase(sessions domain.SessionRepository, log *slog.Logger) *LogoutUseCase {
	return &LogoutUseCase{
		sessions,
		log,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, session_id uint) error {
	return uc.sessions.Delete(ctx, session_id)
}
