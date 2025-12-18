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
	GetDomainGuests(ctx context.Context, domainId uint) ([]Guest, error)
	CreateGuests(ctx context.Context, guests *[]Guest) error
}

type GuestSessionRepository interface {
	Create(ctx context.Context, session *GuestSession) error
	GetCountActiveSessions(ctx context.Context, domain_id uint) (int64, error)
	SetLastActive(ctx context.Context, session_ids map[uint]struct{}, last_active time.Time) error
	GetStaleSessions(ctx context.Context, limit int) (*[]GuestSession, error)
	CloseSessions(ctx context.Context, session_ids []uint) error
	CreateSessions(ctx context.Context, sessions *[]GuestSession) error
}

type DomainRepository interface {
	ByURL(ctx context.Context, url string) (*Domain, error)
	GetCountDomainGuests(ctx context.Context, domainid uint) (int64, error)
	AddDomain(ctx context.Context, d Domain) error
}
