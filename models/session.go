package models

import "time"

type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `json:"user_id"`
	RefreshToken string    `json:"-"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at"`
}
