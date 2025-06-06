package controllers

import (
	"defdrive/models"
	"net/http"
	"os"
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

	// Get username from database
	var user models.User
	if err := fc.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Create user directory if it doesn't exist
	userDir := filepath.Join("/app/data/uploads", user.Username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user directory"})
		return
	}

	// Check if file already exists
	filePath := filepath.Join(userDir, filepath.Base(file.Filename))
	if _, err := os.Stat(filePath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A file with this name already exists in your folder"})
		return
	}

	// Save file to disk
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Store only the relative path (username/filename) in the database
	relativePath := filepath.Join(user.Username, filepath.Base(file.Filename))

	// Create file record in database
	fileRecord := models.File{
		Name:     file.Filename,
		Location: relativePath,
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
}

// ListFiles returns all files belonging to the current user
func (fc *FileController) ListFiles(c *gin.Context) {
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
}

// TogglePublicAccess changes the public status of a file
func (fc *FileController) TogglePublicAccess(c *gin.Context) {
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
}

// DeleteFile removes a file from the system
func (fc *FileController) DeleteFile(c *gin.Context) {
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

	// Reconstruct the full file path to delete the physical file
	fullPath := filepath.Join("/app/data/uploads", file.Location)

	// Delete the file record
	if err := fc.DB.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}

	// Delete the physical file from the host system
	if err := os.Remove(fullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the physical file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
