package analytics

import (
	"context"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/infrastructure/jwt"
)

type CollectRecordEventsUseCase struct {
	events domain.RecordEventRepository
	jwt    *jwt.JWTProvider
}

func NewCollectRecordEventsUseCase(events domain.RecordEventRepository, jwt *jwt.JWTProvider) *CollectRecordEventsUseCase {
	return &CollectRecordEventsUseCase{events, jwt}
}

func (uc *CollectRecordEventsUseCase) Execute(ctx context.Context, events *[]domain.RecordEvent, session_id uint) error {
	//указываем id сессии в ивентах
	for _, event := range *events {
		event.SessionID = session_id
	}

	if err := uc.events.SaveEvents(ctx, events); err != nil {
		return err
	}

	return nil
}
