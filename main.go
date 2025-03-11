package main

import (
	"defdrive/models"
	"defdrive/routes"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using environment variables")
	}

	// Get database URL from environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL environment variable not set")
	}

	// Log connection attempt
	log.Printf("Connecting to database using DATABASE_URL")

	// Connect to the database with retry logic
	var db *gorm.DB
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	// Ensure the tables are created in the correct order
	err = db.AutoMigrate(&models.User{}, &models.File{}, &models.Access{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully")

	// Set up router
	router := routes.SetupRouter(db)

	// Get server port from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	// Start the server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
