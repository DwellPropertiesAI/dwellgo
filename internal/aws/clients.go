package aws

import (
	"context"
	"fmt"

	"dwell/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Clients struct {
	Cognito *cognitoidentityprovider.Client
	S3      *s3.Client
	Bedrock *bedrockruntime.Client
	SNS     *sns.Client
	SES     *ses.Client
}

func NewClients(cfg *config.AWSConfig) (*Clients, error) {
	// Load AWS configuration
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Initialize Cognito client
	cognitoClient := cognitoidentityprovider.NewFromConfig(awsCfg)

	// Initialize S3 client
	s3Client := s3.NewFromConfig(awsCfg)

	// Initialize Bedrock client
	bedrockClient := bedrockruntime.NewFromConfig(awsCfg)

	// Initialize SNS client
	snsClient := sns.NewFromConfig(awsCfg)

	// Initialize SES client
	sesClient := ses.NewFromConfig(awsCfg)

	return &Clients{
		Cognito: cognitoClient,
		S3:      s3Client,
		Bedrock: bedrockClient,
		SNS:     snsClient,
		SES:     sesClient,
	}, nil
}

// GetCognitoClient returns the Cognito client
func (c *Clients) GetCognitoClient() *cognitoidentityprovider.Client {
	return c.Cognito
}

// GetS3Client returns the S3 client
func (c *Clients) GetS3Client() *s3.Client {
	return c.S3
}

// GetBedrockClient returns the Bedrock client
func (c *Clients) GetBedrockClient() *bedrockruntime.Client {
	return c.Bedrock
}

// GetSNSClient returns the SNS client
func (c *Clients) GetSNSClient() *sns.Client {
	return c.SNS
}

// GetSESClient returns the SES client
func (c *Clients) GetSESClient() *ses.Client {
	return c.SES
}

