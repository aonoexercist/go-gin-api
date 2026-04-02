package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "s3cr3t-p@ss"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Fatalf("CheckPassword failed for correct password")
	}

	if CheckPassword("wrong-password", hash) {
		t.Fatalf("CheckPassword returned true for wrong password")
	}
}

func TestGenerateTokensClaims(t *testing.T) {
	// ensure deterministic signing key for tests
	jwtKey = []byte("test-secret-key")

	// Access token
	access, err := GenerateAccessToken(42)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}

	token, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("parsed access token invalid: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("access token claims not MapClaims")
	}

	if claims["user_id"] == nil {
		t.Fatalf("access token missing user_id claim")
	}

	// Refresh token
	refresh, err := GenerateRefreshToken(7)
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}

	rtoken, err := jwt.Parse(refresh, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !rtoken.Valid {
		t.Fatalf("parsed refresh token invalid: %v", err)
	}

	rclaims, ok := rtoken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("refresh token claims not MapClaims")
	}

	if rclaims["session_id"] == nil {
		t.Fatalf("refresh token missing session_id claim")
	}

	// verify expirations are in the future
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			t.Fatalf("access token already expired")
		}
	} else {
		t.Fatalf("access token exp claim missing or wrong type")
	}

	if exp, ok := rclaims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			t.Fatalf("refresh token already expired")
		}
	} else {
		t.Fatalf("refresh token exp claim missing or wrong type")
	}
}
