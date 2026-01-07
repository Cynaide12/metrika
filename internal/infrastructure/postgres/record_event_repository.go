package postgres

import (
	"context"
	"errors"
	domain "metrika/internal/domain/analytics"

	"gorm.io/gorm"
)

type RecordEventRepository struct {
	db *gorm.DB
}

func NewRecordEventRepository(db *gorm.DB) *RecordEventRepository {
	return &RecordEventRepository{db}
}

func (r *RecordEventRepository) SaveEvents(ctx context.Context, events *[]domain.RecordEvent) error {
	db := getDB(ctx, r.db)

	var mEvents []RecordEvent
	for _, event := range *events {
		mEvents = append(mEvents, RecordEvent{
			Data:      event.Data,
			Timestamp: event.Timestamp,
			Type:      event.Type,
			SessionID: event.SessionID,
		})
	}

	if err := db.Create(&mEvents).Error; err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated){
			return domain.ErrSessionsNotFound
		}
		return err
	}

	return nil
}

func (r *RecordEventRepository) GetBySessionId(ctx context.Context, session_id uint) (*[]domain.RecordEvent, error) {
	db := getDB(ctx, r.db)

	var mEvents []RecordEvent

	if err := db.Model(&RecordEvent{}).Where("session_id=?", session_id).Order("BY timestamp ASC").Find(&mEvents).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRecordEventsNotFound
		}
		return nil, err
	}

	var dEvents []domain.RecordEvent
	for _, event := range mEvents {
		dEvents = append(dEvents, domain.RecordEvent{
			Type:      event.Type,
			Timestamp: event.Timestamp,
			SessionID: event.SessionID,
			ID:        event.ID,
			Data:      event.Data,
		})
	}

	return &dEvents, nil
}
