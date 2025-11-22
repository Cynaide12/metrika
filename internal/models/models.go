package models

import "time"

type Event struct {
	SessionID string                 `gorm:"column:session_id;NOT NULL"`
	Type      string                 `gorm:"column:type;NOT NULL"`
	UserID    string                 `gorm:"column:user_id;NOT NULL"`
	PageURL   string                 `gorm:"column:page_url;NOT NULL"`
	Element   string                 `gorm:"column:element;NOT NULL"`
	Timestamp time.Time              `gorm:"column:timestamp;NOT NULL"`
	Data      map[string]interface{} `gorm:"column:data;NOT NULL"`
}
