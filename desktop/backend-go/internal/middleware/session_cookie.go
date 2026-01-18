package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SetSessionCookie sets the Better Auth session cookie with environment-dependent configuration
// This centralizes the duplicate cookie-setting logic across all auth handlers
func SetSessionCookie(c *gin.Context, token string) {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	sameSite := http.SameSiteLaxMode // Secure default for production
	secure := isProduction

	// Allow cross-origin cookies in development (different ports)
	// or when explicitly enabled
	if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
		sameSite = http.SameSiteNoneMode
	}

	// For development, use SameSite=None to allow cross-origin cookies
	// Browsers allow SameSite=None without Secure for localhost
	if !isProduction {
		sameSite = http.SameSiteNoneMode
		secure = false // localhost is exempt from Secure requirement for SameSite=None
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    token,
		Path:     "/",
		Domain:   domain,
		MaxAge:   60 * 60 * 24 * 30, // 30 days - persistent login
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// ClearSessionCookie removes the Better Auth session cookie with environment-dependent configuration
// This must match the configuration used when setting the cookie
func ClearSessionCookie(c *gin.Context) {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain == "" {
		domain = "" // Current domain
	}

	sameSite := http.SameSiteLaxMode // Secure default for production
	if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
}
