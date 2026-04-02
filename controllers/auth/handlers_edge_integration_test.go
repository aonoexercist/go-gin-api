package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-gin-api/config"
	"go-gin-api/controllers/auth"
	"go-gin-api/middleware"
	"go-gin-api/models"
	"go-gin-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// freshEdgeDB opens a uniquely-named in-memory SQLite DB per test, migrates all
// required tables, injects it as config.DB, sets both JWT keys to a fixed test
// secret, and returns a ready Gin engine.
func freshEdgeDB(t *testing.T) (*gorm.DB, *gin.Engine) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Permission{}, &models.Role{}, &models.User{}, &models.Session{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	config.DB = db
	auth.SetJWTKey([]byte("edge-secret"))
	middleware.SetJWTKey([]byte("edge-secret"))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	routes.SetupRoutes(r)
	return db, r
}

// jsonBody serialises v to a *bytes.Reader suitable for an http.Request body.
func jsonBody(t *testing.T, v interface{}) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewReader(b)
}

// postJSON fires a POST with JSON content-type and returns the recorder.
func postJSON(r *gin.Engine, path string, body interface{}, t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// getCookieValue extracts a named cookie value from a response recorder.
func getCookieValue(w *httptest.ResponseRecorder, name string) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// ─── Register edge cases ──────────────────────────────────────────────────────

// TestRegister_MissingName — binding validation: name is required → 400.
func TestRegister_MissingName(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := postJSON(r, "/auth/register", map[string]string{
		"email": "noname@example.com", "password": "pass1234",
	}, t)
	if w.Code != 400 {
		t.Fatalf("expected 400 for missing name, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestRegister_InvalidEmail — binding validation: invalid email format → 400.
func TestRegister_InvalidEmail(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := postJSON(r, "/auth/register", map[string]string{
		"name": "Bad", "email": "not-an-email", "password": "pass1234",
	}, t)
	if w.Code != 400 {
		t.Fatalf("expected 400 for invalid email, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestRegister_PasswordTooShort — binding validation: password min=6 → 400.
func TestRegister_PasswordTooShort(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := postJSON(r, "/auth/register", map[string]string{
		"name": "Short", "email": "short@example.com", "password": "12345",
	}, t)
	if w.Code != 400 {
		t.Fatalf("expected 400 for short password, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestRegister_DuplicateEmail — second registration with the same email → 500.
func TestRegister_DuplicateEmail(t *testing.T) {
	_, r := freshEdgeDB(t)
	payload := map[string]string{"name": "Dup", "email": "dup@example.com", "password": "pass1234"}

	for i, want := range []int{200, 500} {
		w := postJSON(r, "/auth/register", payload, t)
		if w.Code != want {
			t.Fatalf("attempt %d: expected %d, got %d body=%s", i+1, want, w.Code, w.Body.String())
		}
	}
}

// ─── Login edge cases ─────────────────────────────────────────────────────────

// TestLogin_WrongPassword — correct email but wrong password → 401.
func TestLogin_WrongPassword(t *testing.T) {
	_, r := freshEdgeDB(t)

	postJSON(r, "/auth/register", map[string]string{
		"name": "Pw", "email": "pw@example.com", "password": "correct1234",
	}, t)

	w := postJSON(r, "/auth/login", map[string]string{
		"email": "pw@example.com", "password": "wrongpassword",
	}, t)
	if w.Code != 401 {
		t.Fatalf("expected 401 for wrong password, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestLogin_NonExistentEmail — email that was never registered → 401.
func TestLogin_NonExistentEmail(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := postJSON(r, "/auth/login", map[string]string{
		"email": "ghost@example.com", "password": "doesntmatter",
	}, t)
	if w.Code != 401 {
		t.Fatalf("expected 401 for unknown email, got %d body=%s", w.Code, w.Body.String())
	}
}

// ─── /services/me edge cases ──────────────────────────────────────────────────

// TestMe_NoToken — calling /services/me without a cookie → 401 missing_token.
func TestMe_NoToken(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/services/me", nil)
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401 for missing token on /services/me, got %d", w.Code)
	}
}

// TestMe_ExpiredToken — access_token with past expiry → 401 token_expired.
func TestMe_ExpiredToken(t *testing.T) {
	_, r := freshEdgeDB(t)

	expiredToken := buildExpiredToken(t, []byte("edge-secret"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/services/me", nil)
	req.Header.Set("Cookie", "access_token="+expiredToken)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for expired token, got %d body=%s", w.Code, w.Body.String())
	}
}

// ─── Refresh edge cases ───────────────────────────────────────────────────────

// TestRefresh_NoCookie — /auth/refresh with no cookie → 401.
func TestRefresh_NoCookie(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401 for missing refresh cookie, got %d", w.Code)
	}
}

// TestRefresh_MalformedToken — garbage refresh_token cookie → 401.
func TestRefresh_MalformedToken(t *testing.T) {
	_, r := freshEdgeDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.Header.Set("Cookie", "refresh_token=garbage.token.value")
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401 for malformed refresh token, got %d", w.Code)
	}
}

// ─── Full refresh → logout flow ───────────────────────────────────────────────

// TestRefreshAndLogout_EndToEnd — register → login → refresh → logout.
func TestRefreshAndLogout_EndToEnd(t *testing.T) {
	db, r := freshEdgeDB(t)

	// 1. Register
	if w := postJSON(r, "/auth/register", map[string]string{
		"name": "FlowUser", "email": "flow@example.com", "password": "flowpass123",
	}, t); w.Code != 200 {
		t.Fatalf("register: code=%d body=%s", w.Code, w.Body.String())
	}

	// 2. Login → capture refresh_token
	loginW := postJSON(r, "/auth/login", map[string]string{
		"email": "flow@example.com", "password": "flowpass123",
	}, t)
	if loginW.Code != 200 {
		t.Fatalf("login: code=%d body=%s", loginW.Code, loginW.Body.String())
	}
	refreshToken := getCookieValue(loginW, "refresh_token")
	if refreshToken == "" {
		t.Fatalf("refresh_token cookie not set on login")
	}

	// 3. Refresh → should rotate token and return 200
	refreshW := httptest.NewRecorder()
	refreshReq := httptest.NewRequest("POST", "/auth/refresh", nil)
	refreshReq.Header.Set("Cookie", "refresh_token="+refreshToken)
	r.ServeHTTP(refreshW, refreshReq)
	if refreshW.Code != 200 {
		t.Fatalf("refresh: code=%d body=%s", refreshW.Code, refreshW.Body.String())
	}

	newRefreshToken := getCookieValue(refreshW, "refresh_token")
	if newRefreshToken == "" {
		t.Fatalf("new refresh_token not set after /auth/refresh")
	}

	// 4. Logout using the NEW token
	logoutW := httptest.NewRecorder()
	logoutReq := httptest.NewRequest("POST", "/auth/logout", nil)
	logoutReq.Header.Set("Cookie", "refresh_token="+newRefreshToken)
	r.ServeHTTP(logoutW, logoutReq)
	if logoutW.Code != 200 {
		t.Fatalf("logout: code=%d body=%s", logoutW.Code, logoutW.Body.String())
	}

	// 5. Session must no longer exist in DB
	var count int64
	db.Model(&models.Session{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected all sessions deleted after logout, found %d", count)
	}
}

// ─── Admin RBAC via HTTP ──────────────────────────────────────────────────────

// TestAdminRoute_ForbiddenWithoutRole — authenticated user without super_admin role → 403.
func TestAdminRoute_ForbiddenWithoutRole(t *testing.T) {
	_, r := freshEdgeDB(t)

	postJSON(r, "/auth/register", map[string]string{
		"name": "PlainUser", "email": "plain@example.com", "password": "plainpass123",
	}, t)

	loginW := postJSON(r, "/auth/login", map[string]string{
		"email": "plain@example.com", "password": "plainpass123",
	}, t)
	accessToken := getCookieValue(loginW, "access_token")
	if accessToken == "" {
		t.Fatalf("access_token not set after login")
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/users/", nil)
	req.Header.Set("Cookie", "access_token="+accessToken)
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected 403 for user without super_admin role, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestAdminRoute_Unauthenticated — no token at all on an admin route → 401.
func TestAdminRoute_Unauthenticated(t *testing.T) {
	_, r := freshEdgeDB(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/users/", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated admin access, got %d", w.Code)
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// buildExpiredToken creates an already-expired HS256 JWT signed with key.
func buildExpiredToken(t *testing.T, key []byte) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(-10 * time.Minute).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("buildExpiredToken: %v", err)
	}
	return s
}
