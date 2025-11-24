package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth creates a middleware that validates X-API-Key header
func APIKeyAuth(validAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract X-API-Key from request header
		apiKey := c.GetHeader("X-API-Key")

		// Check if API key is missing
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			c.Abort()
			return
		}

		// Check if API key is invalid
		if apiKey != validAPIKey {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// API key is valid, continue with the request
		c.Next()
	}
}
