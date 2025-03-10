package controllers

import (
	"crypto/sha256"
	"defdrive/models"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileController struct {
	DB *gorm.DB
}

func NewFileController(db *gorm.DB) *FileController {
	return &FileController{DB: db}
}

// Upload handles file uploads
func (fc *FileController) Upload(c *gin.Context) {
	// Get the logged-in user from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Find the user in the database
	var user models.User
	if result := fc.DB.Where("user_id = ?", userID).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get the file from the form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Generate a unique file ID
	fileID := uuid.New().String()

	// Create the directory if it doesn't exist
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "./data"
	}

	// Create a directory specific to the user
	userDirPath := filepath.Join(dataPath, user.UserID)
	if err := os.MkdirAll(userDirPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory"})
		return
	}

	// Use the fileID in the filename to ensure uniqueness
	// This way each file will have a unique name even if uploaded by the same user
	filename := fmt.Sprintf("%s-%s", fileID, file.Filename)
	filePath := filepath.Join(userDirPath, filename)

	// Save the file
	if err := saveFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Calculate SHA-256 hash of the file
	fileHash, err := calculateSHA256(filePath)
	if err != nil {
		// If hash calculation fails, clean up and return error
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate file hash"})
		return
	}

	// Get public status from form, default to false (private)
	public := false
	if publicStr := c.PostForm("public"); publicStr == "true" {
		public = true
	}

	// Create the file record in the database
	fileRecord := models.File{
		FileID:   fileID,
		Name:     file.Filename, // Keep the original filename in the database
		Location: filePath,      // Store the full path with unique filename
		OwnerID:  user.UserID,   // Set foreign key directly
		Public:   public,        // Using the boolean field
		Size:     file.Size,
		Hash:     fileHash,
		Accesses: []models.Access{}, // Empty initially
	}

	if result := fc.DB.Create(&fileRecord); result.Error != nil {
		// Clean up the file if the database operation fails
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"file":    fileRecord,
	})
}

// ListFiles retrieves all files for the authenticated user
func (fc *FileController) ListFiles(c *gin.Context) {
	// Get the logged-in user from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var files []models.File
	if result := fc.DB.Where("owner_id = ?", userID).Find(&files); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// TogglePublicAccess allows users to change a file's public status
func (fc *FileController) TogglePublicAccess(c *gin.Context) {
	// Get the logged-in user from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get file ID from URL parameter
	fileID := c.Param("fileID")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// Find the file and check if user has permission
	var file models.File
	if result := fc.DB.Where("file_id = ? AND owner_id = ?", fileID, userID).First(&file); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or you don't have permission"})
		return
	}

	// Parse request body
	var input struct {
		Public bool `json:"public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the file's public status
	file.Public = input.Public
	if result := fc.DB.Save(&file); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File access updated successfully",
		"file":    file,
	})
}

// DeleteFile completely removes a file, including database records and the physical file
func (fc *FileController) DeleteFile(c *gin.Context) {
	// Get the logged-in user from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get file ID from URL parameter
	fileID := c.Param("fileID")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// Find the file and check if the user has permission
	var file models.File
	if result := fc.DB.Where("file_id = ? AND owner_id = ?", fileID, userID).First(&file); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or you don't have permission"})
		return
	}

	// Start a transaction to ensure database consistency
	tx := fc.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	// Delete all access records for this file first
	if err := tx.Where("file_id = ?", fileID).Delete(&models.Access{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete access records"})
		return
	}

	// Delete the file record from the database
	if err := tx.Delete(&file).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Delete the physical file
	if err := os.Remove(file.Location); err != nil {
		// Note: We don't rollback database changes if physical file deletion fails
		// This is a design decision - you might want to handle this differently
		c.JSON(http.StatusOK, gin.H{
			"message": "File record deleted but could not delete physical file",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

// Helper function to save a file
func saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// calculateSHA256 computes the SHA-256 hash of a file
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
