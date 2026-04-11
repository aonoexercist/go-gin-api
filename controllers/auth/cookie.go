package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func cookieConfig() (domain string, secure bool, sameSite http.SameSite) {
	if os.Getenv("GIN_MODE") == "release" {
		domain = ".theworkpc.com"
	} else {
		domain = os.Getenv("COOKIE_DOMAIN") // empty = current host only (safe for localhost)
	}
	secure = os.Getenv("GIN_MODE") == "release"
	if secure {
		sameSite = http.SameSiteNoneMode // cross-origin in prod (requires Secure=true)
	} else {
		sameSite = http.SameSiteLaxMode // dev: no Secure required
	}
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
