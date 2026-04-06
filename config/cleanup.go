package config

import (
	"go-gin-api/models"
	"log"
	"time"

	"gorm.io/gorm"
)

func CleanupSessions(db *gorm.DB) {
	// Deletes all records where the expiry time is in the past
	result := db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})

	if result.Error != nil {
		log.Printf("Failed to cleanup sessions: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d expired sessions", result.RowsAffected)
	}
}
