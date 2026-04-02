package auth

import (
	"testing"

	"go-gin-api/config"
	"go-gin-api/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupMemoryDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite memory: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// attach to global config for functions that read config.DB
	config.DB = db
	return db
}

func TestFindOrCreateUser_ByGoogleID(t *testing.T) {
	db := setupMemoryDB(t)

	gid := "google-123"
	u := models.User{Email: "a@example.com", Name: "A", GoogleID: &gid}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	info := models.GoogleUserInfo{ID: gid, Email: "a@example.com", Name: "A"}
	got, err := FindOrCreateUser(db, info)
	if err != nil {
		t.Fatalf("FindOrCreateUser returned error: %v", err)
	}

	if got.ID != u.ID {
		t.Fatalf("expected existing user id %d, got %d", u.ID, got.ID)
	}
}

func TestFindOrCreateUser_ByEmail_LinkGoogle(t *testing.T) {
	db := setupMemoryDB(t)

	u := models.User{Email: "b@example.com", Name: "B"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	info := models.GoogleUserInfo{ID: "google-999", Email: "b@example.com", Name: "B"}
	got, err := FindOrCreateUser(db, info)
	if err != nil {
		t.Fatalf("FindOrCreateUser returned error: %v", err)
	}

	if got.GoogleID == nil || *got.GoogleID != "google-999" {
		t.Fatalf("expected GoogleID to be linked, got %+v", got.GoogleID)
	}
}

func TestFindOrCreateUser_CreateNew(t *testing.T) {
	db := setupMemoryDB(t)

	info := models.GoogleUserInfo{ID: "g-new", Email: "new@example.com", Name: "New"}
	got, err := FindOrCreateUser(db, info)
	if err != nil {
		t.Fatalf("FindOrCreateUser error: %v", err)
	}

	if got.ID == 0 {
		t.Fatalf("expected created user to have ID set")
	}
}
