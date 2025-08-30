package services

import (
	"dwell/internal/aws"
	"dwell/internal/config"
	"dwell/internal/database"
)

// Services holds all service instances
type Services struct {
	authService *AuthService
	aiService   *AIService
	s3Service   *S3Service
	// Add other services as they are implemented
}

// NewServices creates and returns a new Services instance
func NewServices(cfg *config.Config, db *database.Connection) *Services {
	// Initialize AWS clients
	awsClients, err := aws.NewClients(&cfg.AWS)
	if err != nil {
		panic(err) // This should be handled more gracefully in production
	}

	// Initialize individual services
	authService := NewAuthService(awsClients, cfg)
	aiService := NewAIService(awsClients, cfg)
	s3Service := NewS3Service(awsClients, cfg)

	return &Services{
		authService: authService,
		aiService:   aiService,
		s3Service:   s3Service,
	}
}

// GetAuthService returns the auth service instance
func (s *Services) GetAuthService() *AuthService {
	return s.authService
}

// GetAIService returns the AI service instance
func (s *Services) GetAIService() *AIService {
	return s.aiService
}

// GetS3Service returns the S3 service instance
func (s *Services) GetS3Service() *S3Service {
	return s.s3Service
}
