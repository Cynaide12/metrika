package analytics

import "time"

type GuestSession struct {
	ID uint
	GuestID   uint
	IPAddress string
	Active    bool
	EndTime   *time.Time
	LastActive time.Time
}