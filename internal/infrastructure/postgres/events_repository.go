package postgres

import (
	"context"
	domain "metrika/internal/domain/analytics"
	"time"

	"gorm.io/gorm"
)

type EventsRepository struct {
	db *gorm.DB
}

func NewEventsRepository(db *gorm.DB) *EventsRepository {
	return &EventsRepository{db}
}

func (d *EventsRepository) SaveEvents(ctx context.Context, events *[]domain.Event) error {
	db := getDB(ctx, d.db)

	var mEvents []Event
	for _, event := range *events {
		mEvents = append(mEvents, Event{
			SessionID: event.SessionID,
			Type:      event.Type,
			Element:   event.Element,
			PageURL:   event.PageURL,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		})
	}

	if err := db.Model(&Event{}).Create(&mEvents).Error; err != nil {
		return err
	}

	var ids []uint
	for _, e := range *events {
		ids = append(ids, e.ID)
	}

	if err := db.Model(&GuestSession{}).Where("active = true AND id IN ?", ids).Updates(&GuestSession{LastActive: time.Now()}).Error; err != nil {
		return err
	}

	return nil
}
