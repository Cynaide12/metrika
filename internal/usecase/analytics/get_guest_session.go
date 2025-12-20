package analytics

import (
	"context"
	"errors"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/pkg/logger/sl"
	"time"
)

type GetGuestSessionUseCase struct {
	guests   domain.GuestsRepository
	sessions domain.GuestSessionRepository
	domains  domain.DomainRepository
	logger   *slog.Logger
}

func NewGetGuestSessionUseCase(
	guests domain.GuestsRepository,
	sessions domain.GuestSessionRepository,
	domain domain.DomainRepository,
	logger *slog.Logger,
) *GetGuestSessionUseCase {
	return &GetGuestSessionUseCase{guests, sessions, domain, logger}
}

func (gc *GetGuestSessionUseCase) Execute(ctx context.Context, FingerprintID, IPAddress string, domainUrl string) (*domain.GuestSession, error) {

	//ищем домен
	dom, err := gc.domains.ByURL(ctx, domainUrl)
	if err != nil {
		if errors.Is(err, domain.ErrDomainNotFound) {
			return nil, err
		}
		gc.logger.Error("ошибка получения домена", sl.Err(err))
		return nil, err
	}

	//ищем или создаем юзера по переданному отпечатку
	guest, err := gc.guests.FirstOrCreate(ctx, FingerprintID, dom.ID)
	if err != nil {
		gc.logger.Error("ошибка получения гостевого юзера по f_id", sl.Err(err))
		return nil, err
	}

	//ищем активную сессию юзера
	activeSession, err := gc.sessions.LastActiveByGuestId(ctx, guest.ID)
	if err != nil && errors.Is(err, domain.ErrLastActiveSessionNotFound){
		gc.logger.Error("ошибка получения последней активной сессии гостевого")
		return nil, err
	}
	//если активная сессия уже есть,
	//то возвращаем ее, не создавая новую
	if err == nil && activeSession != nil{
		return activeSession, nil
	}

	//если активных сессий нет - создаем новую
	session := domain.GuestSession{
		GuestID:    guest.ID,
		IPAddress:  IPAddress,
		EndTime:    nil,
		Active:     true,
		LastActive: time.Now(),
	}

	if err := gc.sessions.Create(ctx, &session); err != nil {
		gc.logger.Error("ошибка создания новой сессии гостю", sl.Err(err))
		return nil, err
	}

	gc.logger.Debug("SESSION", slog.Any("session", session))

	return &session, nil
}
