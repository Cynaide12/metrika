package analytics

import (
	"context"
	"errors"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/pkg/logger/sl"
	"time"
)

type CreateGuestSessionUseCase struct {
	guests   domain.GuestsRepository
	sessions domain.GuestSessionRepository
	domains  domain.DomainRepository
	logger   *slog.Logger
}

func NewCreateGuestSessionUseCase(
	guests domain.GuestsRepository,
	sessions domain.GuestSessionRepository,
	domain domain.DomainRepository,
	logger *slog.Logger,
) *CreateGuestSessionUseCase {
	return &CreateGuestSessionUseCase{guests, sessions, domain, logger}
}

func (gc *CreateGuestSessionUseCase) Execute(ctx context.Context, FingerprintID, IPAddress string, domainUrl string) (*domain.GuestSession, error) {

	//ищем домен
	dom, err := gc.domains.ByURL(ctx, domainUrl)
	if err != nil {
		if errors.Is(err, domain.ErrDomainNotFound) {
			return nil, err
		}
		gc.logger.Error("ошибка получения домена", sl.Err(err))
		return nil, err
	}

	//ищем юзера по переданному отпечатку
	guest, err := gc.guests.FirstOrCreate(ctx, FingerprintID, dom.ID)
	if err != nil {
		gc.logger.Error("ошибка получения юзера по f_id", sl.Err(err))
		return nil, err
	}

	session := domain.GuestSession{
		GuestID:    guest.ID,
		IPAddress:  IPAddress,
		EndTime:    nil,
		Active:     true,
		LastActive: time.Now(),
	}

	if err := gc.sessions.Create(ctx, &session); err != nil {
		gc.logger.Error("ошибка создания новой сессии", sl.Err(err))
		return nil, err
	}

	return &session, nil
}
