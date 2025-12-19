package analytics

import "time"

type Event struct {
	ID        uint
	SessionID uint
	Type      string
	PageURL   string
	Element   string
	Timestamp time.Time
	Data      map[string]any
}
