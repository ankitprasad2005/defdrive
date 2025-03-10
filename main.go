package main

import (
	"defdrive/models"
	"defdrive/routes"
	"fmt"
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

	// Get database connection parameters from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWD") // Changed from DB_PASSWORD to DB_PASSWD to match .env
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	timezone := os.Getenv("TZ")

	// Log connection details for debugging
	log.Printf("Connecting to database: host=%s, user=%s, dbname=%s, port=%s",
		dbHost, dbUser, dbName, dbPort)

	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, timezone)

	// Connect to the database with retry logic
	var db *gorm.DB
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

	// Auto migrate models
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
