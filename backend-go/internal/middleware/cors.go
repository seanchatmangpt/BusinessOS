package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/config"
)

// CORSMiddleware returns CORS middleware configured for Better Auth cookies
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"X-Conversation-Id"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}

	return cors.New(corsConfig)
}
