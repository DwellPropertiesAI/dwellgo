package router

import (
	"dwell/internal/controllers"
	"dwell/internal/middleware"
	"dwell/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(services *services.Services) *gin.Engine {
	// Create Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(corsMiddleware())

	// Add request logging middleware
	r.Use(requestLoggingMiddleware())

	// API version 1 routes
	v1 := r.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", healthCheck)

		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			authController := controllers.NewAuthController(services.GetAuthService())
			auth.POST("/signup", authController.SignUp)
			auth.POST("/confirm", authController.ConfirmSignUp)
			auth.POST("/signin", authController.SignIn)
			auth.POST("/refresh", authController.RefreshToken)
			
			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(services.GetAuthService()))
			{
				authProtected.POST("/signout", authController.SignOut)
				authProtected.GET("/profile", authController.GetProfile)
			}
		}

		// AI Chatbot routes (protected)
		ai := v1.Group("/ai")
		ai.Use(middleware.AuthMiddleware(services.GetAuthService()))
		{
			aiController := controllers.NewAIController(services.GetAIService())
			ai.POST("/query", aiController.QueryAI)
			ai.GET("/tips", aiController.GetPropertyManagementTips)
			ai.GET("/history", aiController.GetAIChatHistory)
			ai.GET("/analytics", aiController.GetAIAnalytics)
		}

		// File management routes (protected)
		files := v1.Group("/files")
		files.Use(middleware.AuthMiddleware(services.GetAuthService()))
		{
			s3Controller := controllers.NewS3Controller(services.GetS3Service())
			files.POST("/upload", s3Controller.UploadFile)
			files.DELETE("/delete", s3Controller.DeleteFile)
			files.GET("/list", s3Controller.ListFiles)
			files.GET("/signed-url", s3Controller.GetSignedURL)
			files.GET("/metadata", s3Controller.GetFileMetadata)
		}

		// Landlord-specific routes (protected, landlord only)
		landlord := v1.Group("/landlord")
		landlord.Use(
			middleware.AuthMiddleware(services.GetAuthService()),
			middleware.RequireLandlord(),
		)
		{
			// TODO: Add landlord controller
			// landlordController := controllers.NewLandlordController(services.GetLandlordService())
			// landlord.GET("/dashboard", landlordController.GetDashboard)
			// landlord.GET("/properties", landlordController.GetProperties)
			// landlord.POST("/properties", landlordController.CreateProperty)
			// landlord.PUT("/properties/:id", landlordController.UpdateProperty)
			// landlord.DELETE("/properties/:id", landlordController.DeleteProperty)
			// landlord.GET("/tenants", landlordController.GetTenants)
			// landlord.GET("/payments", landlordController.GetPayments)
			// landlord.GET("/maintenance", landlordController.GetMaintenanceRequests)
		}

		// Tenant-specific routes (protected, tenant only)
		tenant := v1.Group("/tenant")
		tenant.Use(
			middleware.AuthMiddleware(services.GetAuthService()),
			middleware.RequireTenant(),
		)
		{
			// TODO: Add tenant controller
			// tenantController := controllers.NewTenantController(services.GetTenantService())
			// tenant.GET("/dashboard", tenantController.GetDashboard)
			// tenant.POST("/maintenance", tenantController.CreateMaintenanceRequest)
			// tenant.GET("/maintenance", tenantController.GetMaintenanceRequests)
			// tenant.GET("/payments", tenantController.GetPayments)
		}

		// Shared routes (protected, both landlord and tenant)
		shared := v1.Group("/shared")
		shared.Use(
			middleware.AuthMiddleware(services.GetAuthService()),
			middleware.RequireLandlordOrTenant(),
		)
		{
			// TODO: Add shared controller
			// sharedController := controllers.NewSharedController(services.GetSharedService())
			// shared.GET("/notifications", sharedController.GetNotifications)
			// shared.PUT("/notifications/:id/read", sharedController.MarkNotificationRead)
		}

		// Maintenance routes (protected, both landlord and tenant)
		maintenance := v1.Group("/maintenance")
		maintenance.Use(
			middleware.AuthMiddleware(services.GetAuthService()),
			middleware.RequireLandlordOrTenant(),
		)
		{
			// TODO: Add maintenance controller
			// maintenanceController := controllers.NewMaintenanceController(services.GetMaintenanceService())
			// maintenance.POST("/requests", maintenanceController.CreateRequest)
			// maintenance.GET("/requests", maintenanceController.GetRequests)
			// maintenance.GET("/requests/:id", maintenanceController.GetRequest)
			// maintenance.PUT("/requests/:id", maintenanceController.UpdateRequest)
			// maintenance.POST("/requests/:id/photos", maintenanceController.UploadPhotos)
		}

		// Payment routes (protected, both landlord and tenant)
		payments := v1.Group("/payments")
		payments.Use(
			middleware.AuthMiddleware(services.GetAuthService()),
			middleware.RequireLandlordOrTenant(),
		)
		{
			// TODO: Add payment controller
			// paymentController := controllers.NewPaymentController(services.GetPaymentService())
			// payments.GET("/", paymentController.GetPayments)
			// payments.GET("/:id", paymentController.GetPayment)
			// payments.POST("/", paymentController.CreatePayment)
			// payments.PUT("/:id", paymentController.UpdatePayment)
		}

		// Property routes (protected, both landlord and tenant)
		properties := v1.Group("/properties")
		properties.Use(
			middleware.AuthMiddleware(services.GetAuthService()),
			middleware.RequireLandlordOrTenant(),
		)
		{
			// TODO: Add property controller
			// propertyController := controllers.NewPropertyController(services.GetPropertyService())
			// properties.GET("/", propertyController.GetProperties)
			// properties.GET("/:id", propertyController.GetProperty)
			// properties.POST("/", propertyController.CreateProperty)
			// properties.PUT("/:id", propertyController.UpdateProperty)
			// properties.DELETE("/:id", propertyController.DeleteProperty)
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Dwell Property Management API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
		})
	})

	return r
}

// Middleware functions
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func requestLoggingMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"service": "dwell-api",
		"timestamp": gin.H{
			"unix": time.Now().Unix(),
			"iso":  time.Now().Format(time.RFC3339),
		},
	})
}

