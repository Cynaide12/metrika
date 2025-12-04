package models

import (
	"time"
)

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Event struct {
	Model
	SessionID string    `gorm:"column:session_id;NOT NULL"`
	Type      string    `gorm:"column:type;NOT NULL"`
	UserID    string    `gorm:"column:user_id;NOT NULL"`
	PageURL   string    `gorm:"column:page_url;NOT NULL"`
	Element   string    `gorm:"column:element;NOT NULL"`
	Timestamp time.Time `gorm:"column:timestamp;NOT NULL"`
	// Data      map[string]interface{} `gorm:"column:data;NOT NULL"`
}

type Domain struct {
	Model
	SiteURL string `gorm:"column:site_url;unique;NOT NULL" json:"site_url"`
	Users   []User `gorm:"foreignkey:DomainID;constraint:OnDelete:CASCADE"`
}

type User struct {
	Model
	DomainID    uint          `gorm:"column:domain_id;NOT NULL"`
	Fingerprint string        `gorm:"column:f_id" json:"f_id"`
	Sessions    []UserSession `gorm:"foreignkey:UserID;constraint:OnDelete:CASCADE" json:"sessions;omitempty"`
}

type UserSession struct {
	Model
	UserID    uint   `gorm:"column:user_id;NOT NULL" json:"user_id"`
	IPAddress string `gorm:"column:ip_address;NOT NULL" json:"ip_address"`
	Active    bool   `gorm:"column:active;NOT NULL;default:false"`
}
