package controllers

import (
	"net/http"
	"strconv"
	"time"

	"dwell/internal/middleware"
	"dwell/internal/services"

	"github.com/gin-gonic/gin"
)

type S3Controller struct {
	s3Service *services.S3Service
}

func NewS3Controller(s3Service *services.S3Service) *S3Controller {
	return &S3Controller{
		s3Service: s3Service,
	}
}

// UploadFile handles file uploads to S3
// @Summary Upload file to S3
// @Description Upload a file (image, document) to S3 storage
// @Tags File Management
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Param landlord_id formData string true "Landlord ID"
// @Param category formData string true "File category (maintenance_photo, property_photo, document)"
// @Param entity_id formData string true "ID of the related entity"
// @Param description formData string false "File description"
// @Param is_before_photo formData bool false "For maintenance photos: indicates if this is a before photo"
// @Success 200 {object} services.FileUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/upload [post]
func (c *S3Controller) UploadFile(ctx *gin.Context) {
	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "File upload failed",
			Message: "No file provided or invalid file",
		})
		return
	}

	// Get form data
	landlordID := ctx.PostForm("landlord_id")
	category := ctx.PostForm("category")
	entityID := ctx.PostForm("entity_id")
	description := ctx.PostForm("description")
	isBeforePhoto := ctx.PostForm("is_before_photo") == "true"

	// Validate required fields
	if landlordID == "" || category == "" || entityID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Message: "landlord_id, category, and entity_id are required",
		})
		return
	}

	// Verify user has access to the landlord
	if userClaims.LandlordID == nil || userClaims.LandlordID.String() != landlordID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "You can only upload files for your own landlord account",
		})
		return
	}

	// Create upload request
	uploadReq := &services.FileUploadRequest{
		File:          file,
		LandlordID:    landlordID,
		Category:      category,
		EntityID:      entityID,
		Description:   description,
		IsBeforePhoto: isBeforePhoto,
	}

	// Upload file
	response, err := c.s3Service.UploadFile(ctx, uploadReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "File upload failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteFile handles file deletion from S3
// @Summary Delete file from S3
// @Description Delete a file from S3 storage
// @Tags File Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.FileDeleteRequest true "File deletion request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/delete [delete]
func (c *S3Controller) DeleteFile(ctx *gin.Context) {
	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	var req services.FileDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request data",
			Message: err.Error(),
		})
		return
	}

	// Verify user has access to the landlord
	if userClaims.LandlordID == nil || userClaims.LandlordID.String() != req.LandlordID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "You can only delete files for your own landlord account",
		})
		return
	}

	// Delete file
	err := c.s3Service.DeleteFile(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "File deletion failed",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "File deleted successfully",
	})
}

// ListFiles lists files for a specific landlord and category
// @Summary List files
// @Description List files for a specific landlord, category, and entity
// @Tags File Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param landlord_id query string true "Landlord ID"
// @Param category query string true "File category"
// @Param entity_id query string true "Entity ID"
// @Success 200 {array} services.FileUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/list [get]
func (c *S3Controller) ListFiles(ctx *gin.Context) {
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
	landlordID := ctx.Query("landlord_id")
	category := ctx.Query("category")
	entityID := ctx.Query("entity_id")

	// Validate required parameters
	if landlordID == "" || category == "" || entityID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameters",
			Message: "landlord_id, category, and entity_id are required",
		})
		return
	}

	// Verify user has access to the landlord
	if userClaims.LandlordID == nil || userClaims.LandlordID.String() != landlordID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Access denied",
			Message: "You can only list files for your own landlord account",
		})
		return
	}

	// List files
	files, err := c.s3Service.ListFiles(ctx, landlordID, category, entityID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list files",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, files)
}

// GetSignedURL generates a signed URL for temporary file access
// @Summary Get signed URL
// @Description Generate a signed URL for temporary access to a private file
// @Tags File Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param file_key query string true "S3 file key"
// @Param expires query int false "Expiration time in seconds (default: 3600)"
// @Success 200 {object} SignedURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/signed-url [get]
func (c *S3Controller) GetSignedURL(ctx *gin.Context) {
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
	fileKey := ctx.Query("file_key")
	expiresStr := ctx.DefaultQuery("expires", "3600")

	if fileKey == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing file_key parameter",
			Message: "file_key is required",
		})
		return
	}

	// Parse expiration time
	expires := 3600 // default 1 hour
	if expiresStr != "" {
		if parsed, err := strconv.Atoi(expiresStr); err == nil && parsed > 0 {
			expires = parsed
		}
	}

	// TODO: Verify user has access to the file
	// This would typically involve checking file metadata or database records

	// Generate signed URL
	signedURL, err := c.s3Service.GetSignedURL(ctx, fileKey, time.Duration(expires)*time.Second)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate signed URL",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, SignedURLResponse{
		SignedURL: signedURL,
		ExpiresIn: expires,
		FileKey:   fileKey,
	})
}

// GetFileMetadata retrieves metadata for a specific file
// @Summary Get file metadata
// @Description Get metadata for a specific file in S3
// @Tags File Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param file_key query string true "S3 file key"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/metadata [get]
func (c *S3Controller) GetFileMetadata(ctx *gin.Context) {
	// Get user information from context
	userClaims, exists := middleware.GetUserClaimsFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "User not authenticated",
			Message: "Access token not found",
		})
		return
	}

	// Get file key from query parameter
	fileKey := ctx.Query("file_key")
	if fileKey == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing file_key parameter",
			Message: "file_key is required",
		})
		return
	}

	// TODO: Verify user has access to the file
	// This would typically involve checking file metadata or database records

	// Get file metadata
	metadata, err := c.s3Service.GetFileMetadata(ctx, fileKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get file metadata",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, metadata)
}

// Response types
type SignedURLResponse struct {
	SignedURL string `json:"signed_url"`
	ExpiresIn int    `json:"expires_in"`
	FileKey   string `json:"file_key"`
}
