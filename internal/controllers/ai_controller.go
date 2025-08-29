package controllers

import (
	"net/http"

	"dwell/internal/middleware"
	"dwell/internal/services"

	"github.com/gin-gonic/gin"
)

type AIController struct {
	aiService *services.AIService
}

func NewAIController(aiService *services.AIService) *AIController {
	return &AIController{
		aiService: aiService,
	}
}

// QueryAI handles AI chatbot queries
// @Summary Query AI chatbot
// @Description Ask a question to the AI property management assistant
// @Tags AI Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.AIQueryRequest true "AI query request"
// @Success 200 {object} services.AIQueryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ai/query [post]
func (c *AIController) QueryAI(ctx *gin.Context) {
	var req services.AIQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request data",
			Message: err.Error(),
		})
		return
	}

	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Set user type and landlord ID from context
	req.UserType = userClaims.UserType
	if userClaims.LandlordID != nil {
		req.LandlordID = userClaims.LandlordID.String()
	} else {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Landlord ID required",
			Message: "User must be associated with a landlord",
		})
		return
	}

	// Call AI service
	response, err := c.aiService.QueryAI(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to process AI query",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPropertyManagementTips returns AI-generated tips for property management
// @Summary Get property management tips
// @Description Get AI-generated tips for a specific property management category
// @Tags AI Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category query string true "Category of tips (e.g., tenant_management, maintenance, legal_compliance)"
// @Success 200 {object} PropertyManagementTipsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ai/tips [get]
func (c *AIController) GetPropertyManagementTips(ctx *gin.Context) {
	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Ensure user is a landlord
	if userClaims.UserType != "landlord" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "Only landlords can access property management tips",
		})
		return
	}

	// Get category from query parameter
	category := ctx.Query("category")
	if category == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Category required",
			Message: "Category parameter is required",
		})
		return
	}

	// Get landlord ID
	if userClaims.LandlordID == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Landlord ID required",
			Message: "User must be associated with a landlord",
		})
		return
	}

	// Call AI service
	tips, err := c.aiService.GetPropertyManagementTips(ctx, userClaims.LandlordID.String(), category)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate tips",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, PropertyManagementTipsResponse{
		Category: category,
		Tips:     tips,
		Count:    len(tips),
	})
}

// GetAIChatHistory returns the chat history for the current user
// @Summary Get AI chat history
// @Description Get the user's AI chat conversation history
// @Tags AI Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of messages to return (default: 50)"
// @Param offset query int false "Number of messages to skip (default: 0)"
// @Success 200 {object} AIChatHistoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ai/history [get]
func (c *AIController) GetAIChatHistory(ctx *gin.Context) {
	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Get query parameters
	limit := ctx.DefaultQuery("limit", "50")
	offset := ctx.DefaultQuery("offset", "0")

	// TODO: Implement chat history retrieval from database
	// This would typically involve a repository layer to fetch AI chat messages

	// For now, return empty response
	ctx.JSON(http.StatusOK, AIChatHistoryResponse{
		Messages: []interface{}{},
		Total:    0,
		Limit:    50,
		Offset:   0,
	})
}

// GetAIAnalytics returns analytics about AI usage
// @Summary Get AI usage analytics
// @Description Get analytics about AI chatbot usage for the current user/landlord
// @Tags AI Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period for analytics (day, week, month, year) - default: month"
// @Success 200 {object} AIAnalyticsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ai/analytics [get]
func (c *AIController) GetAIAnalytics(ctx *gin.Context) {
	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Get period from query parameter
	period := ctx.DefaultQuery("period", "month")

	// TODO: Implement AI analytics retrieval from database
	// This would involve aggregating data from AI chat messages

	// For now, return placeholder analytics
	ctx.JSON(http.StatusOK, AIAnalyticsResponse{
		Period:        period,
		TotalQueries:  0,
		TotalTokens:   0,
		TotalCost:     0.0,
		AverageTokens: 0,
		PopularTopics: []string{},
		UsageByDay:    map[string]int{},
	})
}

// Response types
type PropertyManagementTipsResponse struct {
	Category string   `json:"category"`
	Tips     []string `json:"tips"`
	Count    int      `json:"count"`
}

type AIChatHistoryResponse struct {
	Messages []interface{} `json:"messages"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

type AIAnalyticsResponse struct {
	Period        string            `json:"period"`
	TotalQueries  int               `json:"total_queries"`
	TotalTokens   int               `json:"total_tokens"`
	TotalCost     float64           `json:"total_cost"`
	AverageTokens int               `json:"average_tokens"`
	PopularTopics []string          `json:"popular_topics"`
	UsageByDay    map[string]int    `json:"usage_by_day"`
}

