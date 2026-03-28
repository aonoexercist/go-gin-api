package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	// 1. Pull values from environment
	domain := os.Getenv("COOKIE_DOMAIN")
	secure := os.Getenv("COOKIE_SECURE") == "true"

	// Common settings to avoid repetition
	// We use Lax for setting (so links from email/other sites work)
	// and Strict for clearing (as seen in your previous snippet).
	sameSite := http.SameSiteLaxMode

	// Access Token (15 mins)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   900,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	// Refresh Token (7 days)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   604800,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func ClearCookies(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
