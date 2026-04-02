package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// set test key
	jwtKey = []byte("mw-test-key")

	// create token
	claims := jwt.MapClaims{"user_id": 77, "exp": time.Now().Add(5 * time.Minute).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokStr, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "access_token="+tokStr)
	c.Request = req

	handler := AuthMiddleware()
	handler(c)

	if c.IsAborted() {
		t.Fatalf("middleware aborted unexpectedly: code %d body %s", w.Code, w.Body.String())
	}

	uid, exists := c.Get("user_id")
	if !exists {
		t.Fatalf("user_id not set in context")
	}

	if uid.(uint) != 77 {
		t.Fatalf("expected user_id 77, got %v", uid)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	jwtKey = []byte("mw-test-key-2")

	claims := jwt.MapClaims{"user_id": 9, "exp": time.Now().Add(-5 * time.Minute).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokStr, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "access_token="+tokStr)
	c.Request = req

	handler := AuthMiddleware()
	handler(c)

	if !c.IsAborted() {
		t.Fatalf("expected middleware to abort for expired token")
	}

	if w.Header().Get("X-Auth-Error") != "token_expired" {
		t.Fatalf("expected X-Auth-Error=token_expired, got %q", w.Header().Get("X-Auth-Error"))
	}
}
