package analytics

import (
	"context"
	domain "metrika/internal/domain/analytics"
)

type CountActiveSessionsUseCase struct {
	sessions domain.GuestSessionRepository
}

func NewCountActiveSessionsUseCase(
	sessions domain.GuestSessionRepository,
) *CountActiveSessionsUseCase {
	return &CountActiveSessionsUseCase{sessions}
}

func (ec *CountActiveSessionsUseCase) Execute(
	ctx context.Context,
	domain_id uint,
) (int64, error) {
	return ec.sessions.GetCountActiveSessions(ctx, domain_id)
}
