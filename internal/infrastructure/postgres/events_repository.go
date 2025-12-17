package postgres

import (
	"context"
	domain "metrika/internal/domain/analytics"

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

	if err := db.Model(Event{}).Create(&events).Error; err != nil {
		return err
	}

	ids := make(map[uint]struct{})
	for _, e := range *events {
		ids[e.SessionID] = struct{}{}
	}

	// if err := db.Model(&GuestSession{}).Where("active = true AND id IN ?", ids).Updates(&GuestSession{LastActive: time.Now()}).Error; err != nil {
	// 	return err
	// }

	return nil
}
