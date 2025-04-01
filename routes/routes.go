package routes

import (
	"defdrive/controllers"
	"defdrive/middleware"
	"net/http"

	// "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter configures all application routes
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	// router.Use(cors.Default())
	router.Use(middleware.CORSMiddleware())

	// Create controllers
	userController := controllers.NewUserController(db)
	fileController := controllers.NewFileController(db)
	accessController := controllers.NewAccessController(db)
	linkController := controllers.NewLinkController(db)

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
			protected.GET("/accesses/:accessID", accessController.GetAccess)
		}
	}

	// Public access link route with access restrictions middleware
	router.GET("/link/:hash", middleware.AccessRestrictions(db), linkController.HandleAccessLink)
	// router.GET("/link/:hash", linkController.HandleAccessLink)

	// Health check route
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
