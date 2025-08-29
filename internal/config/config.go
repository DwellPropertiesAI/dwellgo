package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Cognito         CognitoConfig
	S3              S3Config
	Bedrock         BedrockConfig
	SNS             SNSConfig
	SES             SESConfig
}

type CognitoConfig struct {
	UserPoolID     string
	ClientID       string
	ClientSecret   string
	Region         string
}

type S3Config struct {
	BucketName string
	Region     string
}

type BedrockConfig struct
{
	Region string
	Model  string
}

type SNSConfig struct {
	Region string
	TopicARN string
}

type SESConfig struct {
	Region string
	FromEmail string
}

type JWTConfig struct {
	SecretKey string
	Expiry    int // in hours
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "dwell"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Cognito: CognitoConfig{
				UserPoolID:   getEnv("COGNITO_USER_POOL_ID", ""),
				ClientID:     getEnv("COGNITO_CLIENT_ID", ""),
				ClientSecret: getEnv("COGNITO_CLIENT_SECRET", ""),
				Region:       getEnv("COGNITO_REGION", "us-east-1"),
			},
			S3: S3Config{
				BucketName: getEnv("S3_BUCKET_NAME", ""),
				Region:     getEnv("S3_REGION", "us-east-1"),
			},
			Bedrock: BedrockConfig{
				Region: getEnv("BEDROCK_REGION", "us-east-1"),
				Model:  getEnv("BEDROCK_MODEL", "anthropic.claude-3-sonnet-20240229-v1:0"),
			},
			SNS: SNSConfig{
				Region:  getEnv("SNS_REGION", "us-east-1"),
				TopicARN: getEnv("SNS_TOPIC_ARN", ""),
			},
			SES: SESConfig{
				Region:    getEnv("SES_REGION", "us-east-1"),
				FromEmail: getEnv("SES_FROM_EMAIL", ""),
			},
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET_KEY", "your-secret-key"),
			Expiry:    24, // 24 hours
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

