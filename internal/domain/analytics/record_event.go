package analytics

import "time"

type RecordEvent struct {
	ID        uint
	SessionID uint
	Type      string
	Timestamp time.Time
	Data      map[string]any
}
