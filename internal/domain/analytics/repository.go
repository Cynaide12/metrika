package analytics

import (
	"context"
	"time"
)

type EventsRepository interface {
	SaveEvents(ctx context.Context, events *[]Event) error
}

var FindGuestAllowedOrders = map[string]bool{
	"first_visit":          true,
	"last_visit":           true,
	"guest_id":             true,
	"total_second_on_site": true,
	"is_online":            true,
	"session_count":        true,
}

type FindGuestsOptions struct {
	DomainID  uint
	StartDate *time.Time
	EndDate   *time.Time
	Limit     *int
	Offset    *int
	Order     *string
	OrderType *string
}

type GuestsRepository interface {
	FirstOrCreate(ctx context.Context, fingerprint string, domain_id uint) (*Guest, error)
	CreateGuests(ctx context.Context, guests *[]Guest) ([]Guest, error)
	Find(ctx context.Context, opts FindGuestsOptions) ([]Guest, int64, error)
	ByID(ctx context.Context, guest_id uint) (*Guest, error) 
}

type GuestSessionRepositoryByRangeDateOptions struct {
	StartDate     *time.Time
	EndDate       *time.Time
	GuestID       *uint
	Limit         *int
	Offset        *int
	WithoutActive *bool
}

type GetVisitsByIntervalOptions struct {
	Start           time.Time
	End             time.Time
	IntervalMinutes int
	IntervalDiviser int
}

type GuestSessionRepository interface {
	Create(ctx context.Context, session *GuestSession) error
	GetCountActiveSessions(ctx context.Context, domain_id uint) (int64, error)
	SetLastActive(ctx context.Context, session_ids []uint, last_active time.Time) error
	GetStaleSessions(ctx context.Context, limit int) (*[]GuestSession, error)
	CloseSessions(ctx context.Context, session_ids []uint) error
	ByRangeDate(ctx context.Context, opts GuestSessionRepositoryByRangeDateOptions) (*[]GuestSession, error)
	GetVisitsByInterval(
		ctx context.Context,
		domain_id uint,
		opts GetVisitsByIntervalOptions,
	) (*[]GuestSessionsByTimeBucket, error)
	CreateSessions(ctx context.Context, sessions *[]GuestSession) ([]GuestSession, error)
	LastActiveByGuestId(ctx context.Context, guest_id uint) (*GuestSession, error)
}

type DomainRepository interface {
	ByURL(ctx context.Context, url string) (*Domain, error)
	AddDomain(ctx context.Context, site_url string) (*Domain, error)
	GetDomainGuests(ctx context.Context, domainId uint) (*[]Guest, error)
	GetDomainGuestsByFingerprints(ctx context.Context, domainId uint, fingerprints []string) (*[]Guest, error)
	GetCountDomainGuests(ctx context.Context, domain_id uint) (int64, error)
}

type RecordEventRepository interface {
	SaveEvents(ctx context.Context, events *[]RecordEvent) error
	GetBySessionId(ctx context.Context, session_id uint) (*[]RecordEvent, error)
}
