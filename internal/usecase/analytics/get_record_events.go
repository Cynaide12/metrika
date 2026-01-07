package analytics

import (
	"context"
	domain "metrika/internal/domain/analytics"
)

type GetRecordEventsUseCase struct {
	events domain.RecordEventRepository
}

func NewGetRecordEventsUseCase(events domain.RecordEventRepository) *GetRecordEventsUseCase {
	return &GetRecordEventsUseCase{events}
}

func (uc *GetRecordEventsUseCase) Execute(ctx context.Context, session_id uint) (*[]domain.RecordEvent, error) {
	return uc.events.GetBySessionId(ctx, session_id)
}
