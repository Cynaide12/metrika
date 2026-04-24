package metrika

import (
	"context"
	domain "metrika/internal/domain/analytics"
)

type SessionsByIntervalUseCase struct {
	sessions domain.GuestSessionRepository
}

func NewSessionsByIntervalUseCase(
	sessions domain.GuestSessionRepository,
) *SessionsByIntervalUseCase {
	return &SessionsByIntervalUseCase{sessions}
}

func (ec *SessionsByIntervalUseCase) Execute(
	ctx context.Context,
	domain_id uint,
	opts domain.GetVisitsByIntervalOptions,
) (*[]domain.GuestSessionsByTimeBucket, error) {
	return ec.sessions.GetVisitsByInterval(ctx, domain_id, opts)
}
