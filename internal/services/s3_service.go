package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"dwell/internal/aws"
	"dwell/internal/config"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	awsClients *aws.Clients
	config     *config.Config
}

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	File          *multipart.FileHeader
	LandlordID    string
	Category      string
	EntityID      string
	Description   string
	IsBeforePhoto bool
}

// FileUploadResponse represents a file upload response
type FileUploadResponse struct {
	FileKey    string    `json:"file_key"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
	Category   string    `json:"category"`
	EntityID   string    `json:"entity_id"`
}

// FileDeleteRequest represents a file deletion request
type FileDeleteRequest struct {
	FileKey    string `json:"file_key" binding:"required"`
	LandlordID string `json:"landlord_id" binding:"required"`
}

// FileListRequest represents a file listing request
type FileListRequest struct {
	LandlordID string `json:"landlord_id" binding:"required"`
	Category   string `json:"category" binding:"required"`
	EntityID   string `json:"entity_id" binding:"required"`
}

// FileListResponse represents a file listing response
type FileListResponse struct {
	Files []FileInfo `json:"files"`
}

// FileInfo represents file information
type FileInfo struct {
	FileKey     string    `json:"file_key"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Category    string    `json:"category"`
	EntityID    string    `json:"entity_id"`
	Description string    `json:"description"`
}

// SignedURLRequest represents a signed URL request
type SignedURLRequest struct {
	FileKey string `json:"file_key" binding:"required"`
	Expires int    `json:"expires"` // in seconds
}

// SignedURLResponse represents a signed URL response
type SignedURLResponse struct {
	SignedURL string `json:"signed_url"`
	ExpiresIn int    `json:"expires_in"`
	FileKey   string `json:"file_key"`
}

func NewS3Service(awsClients *aws.Clients, config *config.Config) *S3Service {
	return &S3Service{
		awsClients: awsClients,
		config:     config,
	}
}

// UploadFile uploads a file to S3
func (s *S3Service) UploadFile(ctx context.Context, req *FileUploadRequest) (*FileUploadResponse, error) {
	// Generate unique file key
	fileKey := s.generateFileKey(req.LandlordID, req.Category, req.EntityID, req.File.Filename)

	// Open file
	file, err := req.File.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload to S3
	_, err = s.awsClients.GetS3Client().PutObject(ctx, &s3.PutObjectInput{
		Bucket:        awssdk.String(s.config.AWS.S3.BucketName),
		Key:           awssdk.String(fileKey),
		Body:          file,
		ContentType:   awssdk.String(req.File.Header.Get("Content-Type")),
		ContentLength: &req.File.Size,
		Metadata: map[string]string{
			"landlord_id":     req.LandlordID,
			"category":        req.Category,
			"entity_id":       req.EntityID,
			"description":     req.Description,
			"is_before_photo": fmt.Sprintf("%t", req.IsBeforePhoto),
			"original_name":   req.File.Filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.config.AWS.S3.BucketName, s.config.AWS.Region, fileKey)

	return &FileUploadResponse{
		FileKey:    fileKey,
		URL:        url,
		Size:       req.File.Size,
		UploadedAt: time.Now(),
		Category:   req.Category,
		EntityID:   req.EntityID,
	}, nil
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, req *FileDeleteRequest) error {
	_, err := s.awsClients.GetS3Client().DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: awssdk.String(s.config.AWS.S3.BucketName),
		Key:    awssdk.String(req.FileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

// ListFiles lists files for a specific landlord, category, and entity
func (s *S3Service) ListFiles(ctx context.Context, landlordID, category, entityID string) ([]FileInfo, error) {
	prefix := s.generateFileKey(landlordID, category, entityID, "")

	result, err := s.awsClients.GetS3Client().ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: awssdk.String(s.config.AWS.S3.BucketName),
		Prefix: awssdk.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files from S3: %w", err)
	}

	var files []FileInfo
	for _, obj := range result.Contents {
		// Extract metadata from key or get object metadata
		fileInfo := FileInfo{
			FileKey: *obj.Key,
			URL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
				s.config.AWS.S3.BucketName, s.config.AWS.Region, *obj.Key),
			Size:       *obj.Size,
			UploadedAt: *obj.LastModified,
			Category:   category,
			EntityID:   entityID,
		}
		files = append(files, fileInfo)
	}

	return files, nil
}

// GetSignedURL generates a signed URL for temporary file access
func (s *S3Service) GetSignedURL(ctx context.Context, fileKey string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.awsClients.GetS3Client())

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: awssdk.String(s.config.AWS.S3.BucketName),
		Key:    awssdk.String(fileKey),
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// GetFileMetadata retrieves metadata for a specific file
func (s *S3Service) GetFileMetadata(ctx context.Context, fileKey string) (map[string]string, error) {
	result, err := s.awsClients.GetS3Client().HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: awssdk.String(s.config.AWS.S3.BucketName),
		Key:    awssdk.String(fileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return result.Metadata, nil
}

// generateFileKey creates a unique file key for S3
func (s *S3Service) generateFileKey(landlordID, category, entityID, filename string) string {
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	return fmt.Sprintf("%s/%s/%s/%s-%s%s",
		landlordID, category, entityID, baseName, timestamp, ext)
}
