package controllers

import (
	"defdrive/models"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileController struct {
	DB *gorm.DB
}

// NewFileController creates a new file controller
func NewFileController(db *gorm.DB) *FileController {
	return &FileController{DB: db}
}

// Upload handles file uploads
func (fc *FileController) Upload(c *gin.Context) {
	go func() {
		// Get current user ID from context (set by auth middleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Get file from request
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
			return
		}

		// Generate storage path
		storagePath := filepath.Join("uploads", filepath.Base(file.Filename))

		// Save file to disk
		if err := c.SaveUploadedFile(file, storagePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Create file record in database
		fileRecord := models.File{
			Name:     file.Filename,
			Location: storagePath,
			UserID:   userID.(uint),
			Size:     file.Size,
			Public:   false, // Default to private
		}

		if result := fc.DB.Create(&fileRecord); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record file in database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
			"file":    fileRecord,
		})
	}()
}

// ListFiles returns all files belonging to the current user
func (fc *FileController) ListFiles(c *gin.Context) {
	go func() {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var files []models.File
		result := fc.DB.Where("user_id = ?", userID).Find(&files)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve files"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"files": files})
	}()
}

// TogglePublicAccess changes the public status of a file
func (fc *FileController) TogglePublicAccess(c *gin.Context) {
	go func() {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		fileID, err := strconv.ParseUint(c.Param("fileID"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
			return
		}

		var file models.File
		if err := fc.DB.First(&file, uint(fileID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// Check if user owns the file
		if file.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this file"})
			return
		}

		// Parse request body
		var requestBody struct {
			Public bool `json:"public"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Update public status
		file.Public = requestBody.Public
		if err := fc.DB.Save(&file).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "File access updated successfully",
			"file":    file,
		})
	}()
}

// DeleteFile removes a file from the system
func (fc *FileController) DeleteFile(c *gin.Context) {
	go func() {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		fileID, err := strconv.ParseUint(c.Param("fileID"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
			return
		}

		var file models.File
		if err := fc.DB.First(&file, uint(fileID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// Check if user owns the file
		if file.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this file"})
			return
		}

		// Delete any associated accesses first (cascade delete)
		if err := fc.DB.Where("file_id = ?", uint(fileID)).Delete(&models.Access{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file accesses"})
			return
		}

		// Delete the file record
		if err := fc.DB.Delete(&file).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
			return
		}

		// In a production system, you would also delete the physical file here
		// os.Remove(file.Location)

		c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
	}()
}
