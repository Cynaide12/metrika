package analytics

import "time"

type Guest struct {
	ID                 uint      `json:"id"`
	DomainID           uint      `json:"domain_id"`
	Fingerprint        string    `json:"f_id"`
	TotalSecondsOnSite string    `json:"total_seconds_on_site"`
	FirstVisit         time.Time `json:"first_visit"`
	LastVisit          time.Time `json:"last_visit"`
	IsOnline           bool      `json:"is_online"`
	SessionsCount      int       `json:"sessions_count"`
}
