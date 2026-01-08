package analytics

type RecordEvent struct {
	ID        uint `json:"id"`
	SessionID uint `json:"session_id"`
	Type      int `json:"type"`
	Timestamp int64 `json:"timestamp"`
	Data      map[string]any `json:"data"`
}
