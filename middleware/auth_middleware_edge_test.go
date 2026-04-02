package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TestAuthMiddleware_NoCookie — no access_token cookie at all → 401 missing_token.
func TestAuthMiddleware_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	AuthMiddleware()(c)

	if !c.IsAborted() {
		t.Fatalf("expected abort for missing cookie")
	}
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err == nil {
		if body["error"] != "missing_token" {
			t.Fatalf("expected error=missing_token, got %v", body["error"])
		}
	}
}

// TestAuthMiddleware_MalformedToken — garbage string in cookie → 401 invalid_token.
func TestAuthMiddleware_MalformedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "access_token=this.is.not.a.jwt")
	c.Request = req

	AuthMiddleware()(c)

	if !c.IsAborted() {
		t.Fatalf("expected abort for malformed token")
	}
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestAuthMiddleware_WrongSigningKey — token signed with a different key → 401 invalid_token.
func TestAuthMiddleware_WrongSigningKey(t *testing.T) {
	jwtKey = []byte("expected-key")

	claims := jwt.MapClaims{"user_id": float64(1), "exp": time.Now().Add(5 * time.Minute).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokStr, err := tok.SignedString([]byte("completely-different-key"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "access_token="+tokStr)
	c.Request = req

	AuthMiddleware()(c)

	if !c.IsAborted() {
		t.Fatalf("expected abort for wrong signing key")
	}
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestAuthMiddleware_MissingUserIDClaim — valid JWT but no user_id claim → 401.
func TestAuthMiddleware_MissingUserIDClaim(t *testing.T) {
	jwtKey = []byte("missing-claim-key")

	claims := jwt.MapClaims{"exp": time.Now().Add(5 * time.Minute).Unix()} // intentionally no user_id
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokStr, _ := tok.SignedString(jwtKey)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "access_token="+tokStr)
	c.Request = req

	AuthMiddleware()(c)

	if !c.IsAborted() {
		t.Fatalf("expected abort for missing user_id claim")
	}
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestAuthMiddleware_AlgorithmNoneAttack — "alg:none" bypass attempt must be rejected.
func TestAuthMiddleware_AlgorithmNoneAttack(t *testing.T) {
	jwtKey = []byte("some-real-key")

	claims := jwt.MapClaims{"user_id": float64(99), "exp": time.Now().Add(5 * time.Minute).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokStr, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign with none: %v", err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "access_token="+tokStr)
	c.Request = req

	AuthMiddleware()(c)

	if !c.IsAborted() {
		t.Fatalf("alg:none token must be rejected by AuthMiddleware")
	}
}
