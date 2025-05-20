package middleware

import (
	"net/http"
	"order-food-api/core/config"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth() gin.HandlerFunc {
	cfg := config.GetConfig()
	return func(c *gin.Context) {
		apiKey := c.GetHeader("api_key")
		if apiKey != cfg.Auth.ApiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
