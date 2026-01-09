package analytics

import "time"

type Guest struct {
	ID          uint
	DomainID    uint
	Fingerprint string
	FirstVisit  time.Time
	LastVisit time.Time
	Sessions []GuestSession
}