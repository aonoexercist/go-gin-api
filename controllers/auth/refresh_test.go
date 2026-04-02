package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"go-gin-api/config"
	"go-gin-api/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func setupDBForAuth(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	config.DB = db
	return db
}

func TestRefresh_RotateTokenHappyPath(t *testing.T) {
	db := setupDBForAuth(t)

	// set deterministic jwtKey
	jwtKey = []byte("test-secret-refresh")

	user := models.User{Email: "u@example.com", Name: "U"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// create session
	session := models.Session{UserID: user.ID, ExpiresAt: time.Now().Add(24 * time.Hour)}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}

	// create refresh token and store on session
	tokenStr, err := GenerateRefreshToken(session.ID)
	if err != nil {
		t.Fatalf("gen refresh token: %v", err)
	}
	session.RefreshToken = tokenStr
	if err := db.Save(&session).Error; err != nil {
		t.Fatalf("save session: %v", err)
	}

	// prepare request with cookie
	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Add("Cookie", "refresh_token="+tokenStr)
	c.Request = req

	Refresh(c)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// ensure session in DB rotated
	var s models.Session
	if err := db.First(&s, session.ID).Error; err != nil {
		t.Fatalf("find session: %v", err)
	}

	if s.RefreshToken == "" {
		t.Fatalf("expected refresh token set, got empty")
	}

	// ensure token is a valid JWT and references the same session id
	parsed, err := jwt.Parse(s.RefreshToken, func(t *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("rotated refresh token invalid: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("rotated token claims not MapClaims")
	}

	if idf, ok := claims["session_id"].(float64); !ok || uint(idf) != session.ID {
		t.Fatalf("rotated token session_id claim mismatch: got %v", claims["session_id"])
	}
}

func TestRefresh_TokenReuseDetected(t *testing.T) {
	db := setupDBForAuth(t)

	jwtKey = []byte("test-secret-reuse")

	user := models.User{Email: "r@example.com", Name: "R"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// session has a different stored token than the one presented
	session := models.Session{UserID: user.ID, ExpiresAt: time.Now().Add(24 * time.Hour), RefreshToken: "stored-token"}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}

	// attacker presents a different token
	attackerToken, err := GenerateRefreshToken(session.ID)
	if err != nil {
		t.Fatalf("generate attacker token: %v", err)
	}

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Add("Cookie", "refresh_token="+attackerToken)
	c.Request = req

	Refresh(c)

	if w.Code != 401 {
		t.Fatalf("expected 401 for reuse detection, got %d", w.Code)
	}

	// ensure sessions for user deleted
	var count int64
	db.Model(&models.Session{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected sessions deleted, found %d", count)
	}
}

func TestLogout_DeletesSessionAndClearsCookies(t *testing.T) {
	db := setupDBForAuth(t)

	jwtKey = []byte("test-secret-logout")

	user := models.User{Email: "l@example.com", Name: "L"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// create session and token
	session := models.Session{UserID: user.ID, ExpiresAt: time.Now().Add(24 * time.Hour)}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}

	tokenStr, err := GenerateRefreshToken(session.ID)
	if err != nil {
		t.Fatalf("gen token: %v", err)
	}
	session.RefreshToken = tokenStr
	if err := db.Save(&session).Error; err != nil {
		t.Fatalf("save session: %v", err)
	}

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Add("Cookie", "refresh_token="+tokenStr)
	c.Request = req

	Logout(c)

	if w.Code != 200 {
		t.Fatalf("expected 200 on logout, got %d", w.Code)
	}

	var s models.Session
	if err := db.First(&s, session.ID).Error; err == nil {
		t.Fatalf("expected session deleted, but found it")
	}
}
