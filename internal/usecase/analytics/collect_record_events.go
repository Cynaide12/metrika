package analytics

import (
	"context"
	domain "metrika/internal/domain/analytics"
)

type CollectRecordEventsUseCase struct {
	events domain.RecordEventRepository
}

func NewCollectRecordEventsUseCase(events domain.RecordEventRepository) *CollectRecordEventsUseCase {
	return &CollectRecordEventsUseCase{events}
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
