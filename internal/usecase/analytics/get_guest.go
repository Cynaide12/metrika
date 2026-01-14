package analytics

import (
	"context"
	"errors"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/pkg/logger/sl"
)

type GetGuestUseCase struct {
	guests domain.GuestsRepository
	log    *slog.Logger
}

func NewGetGuestUseCase(guests domain.GuestsRepository, log *slog.Logger) *GetGuestUseCase {
	return &GetGuestUseCase{guests, log}
}

func (uc *GetGuestUseCase) Execute(ctx context.Context, guest_id uint) (*domain.Guest, error) {
	guest, err := uc.guests.ByID(ctx, guest_id)
	if err != nil {
		if errors.Is(err, domain.ErrGuestNotFound) {
			return nil, err
		}
		uc.log.Error("ошибка при получении гостя из бд", sl.Err(err))
		return nil, err
	}

	return guest, err
}
