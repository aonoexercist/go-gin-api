package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"go-gin-api/config"
	"go-gin-api/controllers/auth"
	"go-gin-api/middleware"
	"go-gin-api/models"
	"go-gin-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupIntegrationDB(t *testing.T) *gorm.DB {
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

func TestRegisterLoginAndMe_EndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// prepare DB and keys
	setupIntegrationDB(t)
	auth.SetJWTKey([]byte("integration-secret"))
	middleware.SetJWTKey([]byte("integration-secret"))

	// build router with routes
	r := gin.New()
	routes.SetupRoutes(r)

	// 1) Register
	payload := map[string]string{"name": "IntUser", "email": "int@example.com", "password": "pass1234"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("register failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 2) Login
	loginPayload := map[string]string{"email": "int@example.com", "password": "pass1234"}
	lb, _ := json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewReader(lb))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("login failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// extract access_token cookie
	var accessCookie string
	for _, c := range w.Result().Cookies() {
		if c.Name == "access_token" {
			accessCookie = c.Value
		}
	}
	if accessCookie == "" {
		t.Fatalf("access_token cookie not set on login")
	}

	// 3) Call /services/me with cookie
	req = httptest.NewRequest("GET", "/services/me", nil)
	req.Header.Set("Cookie", "access_token="+accessCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("/services/me failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// optional: verify returned JSON contains email
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse /services/me response: %v", err)
	}

	if resp["email"] != "int@example.com" {
		t.Fatalf("expected email int@example.com in response, got %v", resp["email"])
	}
}
