package metrika

import (
	"context"
	"log/slog"
	"metrika/internal/domain/analytics"
)

type ActiveSessionsUseCase struct {
	log      *slog.Logger
	sessions analytics.GuestSessionRepository
}

func NewAciveSessionsUseCase(log *slog.Logger, sessions analytics.GuestSessionRepository) *ActiveSessionsUseCase {
	return &ActiveSessionsUseCase{
		log,
		sessions,
	}
}

func (uc *ActiveSessionsUseCase) Execute(ctx context.Context, domain_id uint) (int64, error) {
	count, err := uc.sessions.GetCountActiveSessions(ctx, domain_id)
	if err != nil {
		if err == analytics.ErrSessionsNotFound {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}
