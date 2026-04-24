package postgres

import (
	"metrika/internal/domain/analytics"
	"time"
)

type GuestDTO struct {
	ID                 uint      `gorm:"column:id"`
	DomainID           uint      `gorm:"column:domain_id;NOT NULL"`
	Fingerprint        string    `gorm:"column:f_id"`
	TotalSecondsOnSite string    `gorm:"column:total_seconds_on_site"`
	FirstVisit         time.Time `gorm:"column:first_visit"`
	LastVisit          time.Time `gorm:"column:last_visit"`
	SessionsCount      int       `gorm:"column:sessions_count"`
	IsOnline           bool      `gorm:"column:is_online"`
}

func (d GuestDTO) ToDomain() analytics.Guest {
	return analytics.Guest{
		ID:                 d.ID,
		DomainID:           d.DomainID,
		Fingerprint:        d.Fingerprint,
		TotalSecondsOnSite: d.TotalSecondsOnSite,
		LastVisit:          d.LastVisit,
		FirstVisit:         d.FirstVisit,
		IsOnline:           d.IsOnline,
		SessionsCount:      d.SessionsCount,
	}
}
