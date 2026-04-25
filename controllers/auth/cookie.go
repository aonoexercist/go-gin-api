package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func cookieConfig() (domain string, secure bool, sameSite http.SameSite) {
	domain = "" // Default to current domain
	if envDomain := os.Getenv("COOKIE_DOMAIN"); envDomain != "" {
		domain = envDomain
	}

	sameSite = http.SameSiteLaxMode
	if os.Getenv("GIN_MODE") == "release" {
		sameSite = http.SameSiteNoneMode
	}

	secure = os.Getenv("GIN_MODE") == "release"
	return
}

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	domain, secure, sameSite := cookieConfig()

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
	domain, secure, sameSite := cookieConfig()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}
