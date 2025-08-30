package controllers

import (
	"net/http"

	"dwell/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SignUp handles user registration
func (c *AuthController) SignUp(ctx *gin.Context) {
	var req services.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := c.authService.SignUp(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// ConfirmSignUp handles user registration confirmation
func (c *AuthController) ConfirmSignUp(ctx *gin.Context) {
	var req ConfirmSignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	err := c.authService.ConfirmSignUp(ctx, req.Email, req.ConfirmationCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Confirmation failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "User registration confirmed successfully",
	})
}

// SignIn handles user authentication
func (c *AuthController) SignIn(ctx *gin.Context) {
	var req services.AuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := c.authService.SignIn(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Authentication failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := c.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Token refresh failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// SignOut handles user sign out
func (c *AuthController) SignOut(ctx *gin.Context) {
	// Get user claims from context (set by auth middleware)
	_, exists := ctx.Get("user_claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Extract access token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing authorization header",
			Message: "Authorization header is required",
		})
		return
	}

	// Extract token (remove "Bearer " prefix)
	token := authHeader[7:] // Skip "Bearer "

	err := c.authService.SignOut(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Sign out failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "User signed out successfully",
	})
}

// GetProfile retrieves user profile information
func (c *AuthController) GetProfile(ctx *gin.Context) {
	// Get user claims from context (set by auth middleware)
	claims, exists := ctx.Get("user_claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// For now, return basic profile info
	// In a real implementation, you would fetch this from the database
	userClaims := claims.(map[string]interface{})
	profile := UserProfileResponse{
		UserID:   userClaims["user_id"].(string),
		UserType: userClaims["user_type"].(string),
		// Add more profile fields as needed
	}

	ctx.JSON(http.StatusOK, profile)
}

// Request types
type ConfirmSignUpRequest struct {
	Email            string `json:"email" binding:"required,email"`
	ConfirmationCode string `json:"confirmation_code" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserProfileResponse struct {
	UserID   string `json:"user_id"`
	UserType string `json:"user_type"`
	// Add more profile fields as needed
}

// Response types
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
