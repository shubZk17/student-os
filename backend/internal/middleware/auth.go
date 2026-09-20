package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"studentos/backend/internal/auth"
)

const (
	ContextUserIDKey = "userID"
	ContextEmailKey  = "userEmail"
	ContextRoleKey   = "userRole"
)

// RequireAuth extracts and validates the JWT Bearer token
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format (expected 'Bearer <token>')"})
			return
		}

		claims, err := auth.ValidateAccessToken(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired access token"})
			return
		}

		// Inject verified user metadata into Gin context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextRoleKey, claims.Role)

		c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from context
func GetUserID(c *gin.Context) string {
	if val, exists := c.Get(ContextUserIDKey); exists {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
