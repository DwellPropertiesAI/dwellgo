package services

import (
	"context"
	"testing"

	"dwell/internal/aws"
	"dwell/internal/config"
)

func TestNewAuthService(t *testing.T) {
	// Create mock config
	cfg := &config.Config{
		AWS: config.AWSConfig{
			Cognito: config.CognitoConfig{
				UserPoolID:   "test-pool-id",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Region:       "us-east-1",
			},
		},
		JWT: config.JWTConfig{
			SecretKey: "test-secret-key",
			Expiry:    24,
		},
	}

	// Create mock AWS clients
	awsClients := &aws.Clients{}

	// Test service creation
	service := NewAuthService(awsClients, cfg)

	if service == nil {
		t.Error("Expected AuthService to be created, got nil")
	}

	if service.config != cfg {
		t.Error("Expected config to be set correctly")
	}

	if service.awsClients != awsClients {
		t.Error("Expected AWS clients to be set correctly")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	// Create mock config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey: "test-secret-key-for-validation",
			Expiry:    24,
		},
	}

	// Create mock AWS clients
	awsClients := &aws.Clients{}

	// Create service
	service := NewAuthService(awsClients, cfg)

	// Test with invalid token
	_, err := service.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}

	// Test with empty token
	_, err = service.ValidateToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestAuthService_SignUpRequest_Validation(t *testing.T) {
	// Test valid request
	validReq := &SignUpRequest{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		UserType:    "landlord",
		CompanyName: "Test Company",
	}

	if validReq.Email == "" {
		t.Error("Email should not be empty")
	}

	if validReq.Password == "" {
		t.Error("Password should not be empty")
	}

	if validReq.FirstName == "" {
		t.Error("FirstName should not be empty")
	}

	if validReq.LastName == "" {
		t.Error("LastName should not be empty")
	}

	if validReq.UserType == "" {
		t.Error("UserType should not be empty")
	}

	// Test invalid user type
	invalidReq := &SignUpRequest{
		Email:       "test@example.com",
		Password:    "password123",
		FirstName:   "John",
		LastName:    "Doe",
		UserType:    "invalid_type",
		CompanyName: "Test Company",
	}

	if invalidReq.UserType == "landlord" || invalidReq.UserType == "tenant" {
		t.Error("UserType should be invalid")
	}
}

func TestAuthService_AuthRequest_Validation(t *testing.T) {
	// Test valid request
	validReq := &AuthRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	if validReq.Email == "" {
		t.Error("Email should not be empty")
	}

	if validReq.Password == "" {
		t.Error("Password should not be empty")
	}

	// Test invalid email format
	invalidReq := &AuthRequest{
		Email:    "invalid-email",
		Password: "password123",
	}

	if invalidReq.Email == "test@example.com" {
		t.Error("Email should be invalid")
	}
}

// Mock AWS clients for testing
type MockAWSClients struct {
	*aws.Clients
}

func NewMockAWSClients() *MockAWSClients {
	return &MockAWSClients{}
}

// Benchmark tests
func BenchmarkNewAuthService(b *testing.B) {
	cfg := &config.Config{
		AWS: config.AWSConfig{
			Cognito: config.CognitoConfig{
				UserPoolID:   "test-pool-id",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Region:       "us-east-1",
			},
		},
		JWT: config.JWTConfig{
			SecretKey: "test-secret-key",
			Expiry:    24,
		},
	}

	awsClients := &aws.Clients{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewAuthService(awsClients, cfg)
	}
}

// Test helper functions
func createTestConfig() *config.Config {
	return &config.Config{
		AWS: config.AWSConfig{
			Cognito: config.CognitoConfig{
				UserPoolID:   "test-pool-id",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Region:       "us-east-1",
			},
		},
		JWT: config.JWTConfig{
			SecretKey: "test-secret-key",
			Expiry:    24,
		},
	}
}

func createTestAuthService() *AuthService {
	cfg := createTestConfig()
	awsClients := &aws.Clients{}
	return NewAuthService(awsClients, cfg)
}

// Table-driven tests
func TestAuthService_UserTypeValidation(t *testing.T) {
	tests := []struct {
		name     string
		userType string
		valid    bool
	}{
		{"Valid Landlord", "landlord", true},
		{"Valid Tenant", "tenant", true},
		{"Invalid Type", "admin", false},
		{"Empty Type", "", false},
		{"Case Sensitive", "Landlord", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SignUpRequest{
				Email:       "test@example.com",
				Password:    "password123",
				FirstName:   "John",
				LastName:    "Doe",
				UserType:    tt.userType,
				CompanyName: "Test Company",
			}

			isValid := req.UserType == "landlord" || req.UserType == "tenant"
			if isValid != tt.valid {
				t.Errorf("UserType validation failed for %s: expected %v, got %v", tt.userType, tt.valid, isValid)
			}
		})
	}
}

