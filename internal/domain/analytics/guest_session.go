package analytics

import "time"

type GuestSession struct {
	ID         uint       `json:"id"`
	GuestID    uint       `json:"guest_id"`
	IPAddress  string     `json:"ip_address,omitempty"`
	Active     bool       `json:"active"`
	EndTime    *time.Time `json:"end_time"`
	LastActive time.Time  `json:"last_active"`
}

type GuestSessionsByTimeBucket struct {
	TimeBucket time.Time `json:"time_bucket"`
	Visits     int       `json:"visits"`
	Uniques    int       `json:"uniques"`
}