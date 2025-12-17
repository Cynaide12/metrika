package analytics

import (
	"context"
	"time"
)

type EventsRepository interface {
	SaveEvents(ctx context.Context, events *[]Event) error 
}

type GuestsRepository interface {
	FirstOrCreate(ctx context.Context, fingerprint string, domain_id uint) (*Guest, error)
}

type GuestSessionRepository interface {
	Create(ctx context.Context, session *GuestSession) error
	GetCountActiveSessions(ctx context.Context, domain_id uint) (int64, error)
	SetLastActive(ctx context.Context, session_ids map[uint]struct{}, last_active time.Time) error
	GetStaleSessions(ctx context.Context, limit int) (*[]GuestSession, error)
}

type DomainRepository interface {
	ByURL(ctx context.Context, url string) (*Domain, error)
}
