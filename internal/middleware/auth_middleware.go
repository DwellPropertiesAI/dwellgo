package middleware

import (
	"net/http"
	"strings"

	"dwell/internal/domain"
	"dwell/internal/services"

	"github.com/gin-gonic/gin"
)

const (
	UserClaimsKey = "user_claims"
)

// AuthMiddleware validates JWT tokens and extracts user claims
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Bearer token not provided",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Token validation failed",
			})
			c.Abort()
			return
		}

		// Store user claims in context
		c.Set(UserClaimsKey, claims)
		c.Next()
	}
}

// RequireLandlord middleware ensures the user is a landlord
func RequireLandlord() gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims, exists := GetUserClaimsFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"message": "Access token not found",
			})
			c.Abort()
			return
		}

		if userClaims.UserType != "landlord" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "This endpoint requires landlord privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenant middleware ensures the user is a tenant
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims, exists := GetUserClaimsFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"message": "Access token not found",
			})
			c.Abort()
			return
		}

		if userClaims.UserType != "tenant" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "This endpoint requires tenant privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireLandlordOrTenant middleware ensures the user is either a landlord or tenant
func RequireLandlordOrTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims, exists := GetUserClaimsFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"message": "Access token not found",
			})
			c.Abort()
			return
		}

		if userClaims.UserType != "landlord" && userClaims.UserType != "tenant" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "This endpoint requires landlord or tenant privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserClaimsFromContext extracts user claims from the Gin context
func GetUserClaimsFromContext(c *gin.Context) (*domain.UserClaims, bool) {
	userClaims, exists := c.Get(UserClaimsKey)
	if !exists {
		return nil, false
	}

	claims, ok := userClaims.(*domain.UserClaims)
	if !ok {
		return nil, false
	}

	return claims, true
}
