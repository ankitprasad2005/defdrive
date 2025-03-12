package routes

import (
	"defdrive/controllers"
	"defdrive/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter configures all application routes
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Create controllers
	userController := controllers.NewUserController(db)
	fileController := controllers.NewFileController(db)
	accessController := controllers.NewAccessController(db)

	// Group API routes
	api := router.Group("/api")
	{
		// User routes (public)
		api.POST("/signup", userController.SignUp)
		api.POST("/login", userController.Login)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// File routes
			protected.POST("/upload", fileController.Upload)
			protected.GET("/files", fileController.ListFiles)
			protected.PUT("/files/:fileID/access", fileController.TogglePublicAccess)
			protected.DELETE("/files/:fileID", fileController.DeleteFile)

			// Access routes
			protected.POST("/files/:fileID/accesses", accessController.CreateAccess)
			protected.GET("/files/:fileID/accesses", accessController.ListAccesses)
			protected.PUT("/accesses/:accessID/access", accessController.UpdateAccess)
			protected.DELETE("/accesses/:accessID", accessController.DeleteAccess)
		}
	}

	return router
}
