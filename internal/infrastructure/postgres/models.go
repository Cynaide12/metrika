package postgres

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
	SessionID uint      `gorm:"column:session_id;NOT NULL"`
	Type      string    `gorm:"column:type;NOT NULL"`
	PageURL   string    `gorm:"column:page_url;NOT NULL"`
	Element   string    `gorm:"column:element;NOT NULL"`
	Timestamp time.Time `gorm:"column:timestamp;NOT NULL"`
	// Data      map[string]interface{} `gorm:"column:data;NOT NULL"`
}

type User struct {
	Model
	Name      string `gorm:"column:name;NOT NULL" json:"name"`
	Email     string `gorm:"column:email;NOT NULL;unique" json:"email"`
	Password  []byte `gorm:"column:password;NOT NULL" json:"password,omitempty"`
	LastLogin string `gorm:"column:last_login" json:"last_login,omitempty"`
}

type UserSession struct {
	Model
	UserID       uint   `gorm:"column:user_id;NOT NULL"`
	RefreshToken string `gorm:"column:refresh_token"`
	UserAgent    string `gorm:"column:user_agent"`
}

type Domain struct {
	Model
	SiteURL string  `gorm:"column:site_url;unique;NOT NULL" json:"site_url"`
	Guests  []Guest `gorm:"foreignkey:DomainID;constraint:OnDelete:CASCADE"`
}

type Guest struct {
	Model
	DomainID    uint           `gorm:"column:domain_id;NOT NULL"`
	Fingerprint string         `gorm:"column:f_id" json:"f_id"`
	Sessions    []GuestSession `gorm:"foreignkey:GuestID;constraint:OnDelete:CASCADE" json:"sessions,omitempty"`
}

type GuestSession struct {
	Model
	GuestID    uint       `gorm:"column:guest_id;NOT NULL" json:"guest_id"`
	IPAddress  string     `gorm:"column:ip_address;NOT NULL" json:"ip_address"`
	Active     bool       `gorm:"column:active;NOT NULL;default:false"`
	EndTime    *time.Time `gorm:"column:end_time;default:NULL"`
	LastActive time.Time  `gorm:"column:last_active;NOT NULL;default:CURRENT_TIMESTAMP"`
}
