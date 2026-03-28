package models

import "time"

type Session struct {
	ID           uint
	UserID       uint
	RefreshToken string
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
}
