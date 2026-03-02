package analytics

import (
	"context"
	"errors"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/pkg/pointers"
)

type GetGuestsUseCase struct {
	guests domain.GuestsRepository
	log    *slog.Logger
}

func NewGetGuestsUseCase(guests domain.GuestsRepository, log *slog.Logger) *GetGuestsUseCase {
	return &GetGuestsUseCase{guests, log}
}

func (uc *GetGuestsUseCase) Execute(ctx context.Context, opts domain.FindGuestsOptions) (*[]domain.Guest, int64, error) {

	//max 100
	if opts.Limit == nil {
		opts.Limit = pointers.NewIntPointer(100)
	}

	//default order
	if opts.Order == nil {
		opts.Order = pointers.NewStringPointer("last_visit")
	}

	//default order type
	if opts.OrderType == nil {
		opts.OrderType = pointers.NewStringPointer("ASC")
	}

	guests, total, err := uc.guests.Find(ctx, opts)
	if err != nil && !errors.Is(err, domain.ErrGuestsNotFound) {
		return nil, 0, err
	}

	return &guests, total, err
}
