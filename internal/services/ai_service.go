package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dwell/internal/aws"
	"dwell/internal/config"
	"dwell/internal/domain"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/google/uuid"
)

type AIService struct {
	awsClients *aws.Clients
	config     *config.Config
}

type AIQueryRequest struct {
	Question   string `json:"question" binding:"required"`
	UserType   string `json:"user_type" binding:"required,oneof=landlord tenant"`
	LandlordID string `json:"landlord_id" binding:"required"`
	TenantID   string `json:"tenant_id,omitempty"`
	Context    string `json:"context,omitempty"` // Additional context about the user's situation
}

type AIQueryResponse struct {
	Answer     string  `json:"answer"`
	ModelUsed  string  `json:"model_used"`
	TokensUsed int     `json:"tokens_used"`
	Cost       float64 `json:"cost"`
	Confidence float64 `json:"confidence"`
}

type ClaudeRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []Message `json:"messages"`
	System           string    `json:"system"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	Content []Content `json:"content"`
	Usage   Usage     `json:"usage"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func NewAIService(awsClients *aws.Clients, config *config.Config) *AIService {
	return &AIService{
		awsClients: awsClients,
		config:     config,
	}
}

// QueryAI processes a question through AWS Bedrock and returns an AI-generated answer
func (s *AIService) QueryAI(ctx context.Context, req *AIQueryRequest) (*AIQueryResponse, error) {
	// Prepare the system prompt based on user type
	systemPrompt := s.buildSystemPrompt(req.UserType, req.Context)

	// Prepare the user message
	userMessage := fmt.Sprintf("Question: %s", req.Question)

	// Create Claude request
	claudeReq := &ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1000,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		System: systemPrompt,
	}

	// Convert to JSON
	requestBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call Bedrock
	invokeInput := &bedrockruntime.InvokeModelInput{
		ModelId:     awssdk.String(s.config.AWS.Bedrock.Model),
		Body:        requestBody,
		ContentType: awssdk.String("application/json"),
	}

	result, err := s.awsClients.GetBedrockClient().InvokeModel(ctx, invokeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	// Parse response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(result.Body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract answer
	var answer string
	if len(claudeResp.Content) > 0 {
		answer = claudeResp.Content[0].Text
	}

	// Calculate cost (approximate - actual costs may vary)
	cost := s.calculateCost(claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)

	// Create AI chat message record
	landlordID, _ := uuid.Parse(req.LandlordID)
	var tenantID *uuid.UUID
	if req.TenantID != "" {
		if id, err := uuid.Parse(req.TenantID); err == nil {
			tenantID = &id
		}
	}

	// TODO: Save AI message to database
	// This would typically be done through a repository layer
	_ = &domain.AIChatMessage{
		LandlordID: landlordID,
		TenantID:   tenantID,
		UserType:   req.UserType,
		Question:   req.Question,
		Answer:     answer,
		ModelUsed:  s.config.AWS.Bedrock.Model,
		TokensUsed: claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		Cost:       cost,
	}

	return &AIQueryResponse{
		Answer:     answer,
		ModelUsed:  s.config.AWS.Bedrock.Model,
		TokensUsed: claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		Cost:       cost,
		Confidence: 0.85, // Placeholder - Claude doesn't provide confidence scores
	}, nil
}

// buildSystemPrompt creates a context-aware system prompt for the AI
func (s *AIService) buildSystemPrompt(userType, context string) string {
	basePrompt := `You are DwellAI, an intelligent property management assistant. You help landlords and tenants with property-related questions and issues.

Key Guidelines:
- Always provide helpful, accurate, and professional advice
- Focus on property management, maintenance, legal compliance, and best practices
- If you're unsure about legal matters, recommend consulting with a professional
- Be empathetic and understanding of the user's situation
- Provide actionable advice when possible

Property Management Topics You Can Help With:
- Maintenance requests and scheduling
- Tenant-landlord relationships and communication
- Property inspections and compliance
- Rent collection and payment issues
- Lease agreements and terms
- Property improvements and renovations
- Emergency procedures
- Local property laws and regulations (general guidance only)

IMPORTANT: For legal advice, always recommend consulting with a qualified attorney or legal professional.`

	if userType == "landlord" {
		basePrompt += `

You are speaking with a LANDLORD. Focus on:
- Property management best practices
- Tenant screening and management
- Maintenance coordination
- Legal compliance for landlords
- Financial management and record keeping
- Property value optimization`
	} else if userType == "tenant" {
		basePrompt += `

You are speaking with a TENANT. Focus on:
- Understanding tenant rights and responsibilities
- Maintenance request procedures
- Communication with landlords
- Lease agreement questions
- Emergency situations
- Tenant advocacy and resources`
	}

	if context != "" {
		basePrompt += fmt.Sprintf("\n\nAdditional Context: %s", context)
	}

	return basePrompt
}

// calculateCost estimates the cost of the AI query based on token usage
func (s *AIService) calculateCost(inputTokens, outputTokens int) float64 {
	// Claude 3 Sonnet pricing (approximate, as of 2024)
	// Input: $3.00 per 1M tokens
	// Output: $15.00 per 1M tokens
	inputCost := float64(inputTokens) * 3.00 / 1000000
	outputCost := float64(outputTokens) * 15.00 / 1000000
	return inputCost + outputCost
}

// GetPropertyManagementTips returns AI-generated tips for property management
func (s *AIService) GetPropertyManagementTips(ctx context.Context, landlordID string, category string) ([]string, error) {
	question := fmt.Sprintf("Provide 5 practical tips for %s in property management. Keep each tip concise and actionable.", category)

	req := &AIQueryRequest{
		Question:   question,
		UserType:   "landlord",
		LandlordID: landlordID,
		Context:    fmt.Sprintf("Category: %s", category),
	}

	resp, err := s.QueryAI(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse the response into individual tips
	tips := s.parseTipsFromResponse(resp.Answer)
	return tips, nil
}

// parseTipsFromResponse extracts individual tips from the AI response
func (s *AIService) parseTipsFromResponse(response string) []string {
	// Simple parsing - look for numbered or bulleted items
	lines := strings.Split(response, "\n")
	var tips []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines that start with numbers, bullets, or dashes
		if (len(line) > 2 && (line[0] >= '1' && line[0] <= '9')) ||
			strings.HasPrefix(line, "- ") ||
			strings.HasPrefix(line, "• ") ||
			strings.HasPrefix(line, "* ") {
			tip := strings.TrimPrefix(line, "- ")
			tip = strings.TrimPrefix(tip, "• ")
			tip = strings.TrimPrefix(tip, "* ")
			// Remove leading numbers and dots
			if len(tip) > 0 && tip[0] >= '1' && tip[0] <= '9' {
				if idx := strings.Index(tip, "."); idx != -1 {
					tip = strings.TrimSpace(tip[idx+1:])
				}
			}
			if tip != "" {
				tips = append(tips, tip)
			}
		}
	}

	// If no structured tips found, return the whole response as one tip
	if len(tips) == 0 && response != "" {
		tips = []string{response}
	}

	return tips
}
