package metrika

import (
	"context"
	domain "metrika/internal/domain/analytics"
	"metrika/pkg/pointers"
)

type SessionsByRangeDateUseCase struct {
	sessions domain.GuestSessionRepository
}

func NewSessionsByRangeDateUseCase(
	sessions domain.GuestSessionRepository,
) *SessionsByRangeDateUseCase {
	return &SessionsByRangeDateUseCase{sessions}
}

func (ec *SessionsByRangeDateUseCase) Execute(
	ctx context.Context,
	domain_id uint,
	opts *domain.GuestSessionRepositoryByRangeDateOptions,
) (*[]domain.GuestSession, error) {

	//не больше 1000 сессий за раз можно извлекать
	if opts.Limit == nil || *opts.Limit > 1000 {
		opts.Limit = pointers.NewIntPointer(1000)
	}

	return ec.sessions.ByRangeDate(ctx, *opts)
}
