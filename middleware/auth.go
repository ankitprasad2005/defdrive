package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired is a middleware to authenticate requests
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer format
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Extract the token
		token := headerParts[1]

		// In a real application, verify the JWT token here
		// For now, we'll use a dummy token validation and user ID extraction

		// Dummy validation - in a real app you would verify the JWT
		if token == "invalid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user ID in context - in a real app this would come from the JWT claims
		// This is a placeholder - replace with actual JWT validation logic
		c.Set("userID", uint(1)) // Example user ID

		c.Next()
	}
}
