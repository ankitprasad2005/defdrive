package controllers

import (
	"defdrive/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccessController struct {
	DB *gorm.DB
}

// NewAccessController creates a new access controller
func NewAccessController(db *gorm.DB) *AccessController {
	return &AccessController{DB: db}
}

// CreateAccess generates a new access record for a file
func (ac *AccessController) CreateAccess(c *gin.Context) {
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

	// Check if the user owns the file
	var file models.File
	if err := ac.DB.First(&file, uint(fileID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to create access for this file"})
		return
	}

	// Parse request body
	var accessRequest struct {
		Name    string `json:"name"`
		Subnet  string `json:"subnet"`
		IP      string `json:"ip"`
		Expires string `json:"expires"`
		Public  bool   `json:"public"`
	}

	if err := c.ShouldBindJSON(&accessRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Generate a unique access link
	link := uuid.New().String()

	// Create the access record
	access := models.Access{
		FileID:  uint(fileID),
		Name:    accessRequest.Name,
		Link:    link,
		Subnet:  accessRequest.Subnet,
		IP:      accessRequest.IP,
		Expires: accessRequest.Expires,
		Public:  accessRequest.Public,
	}

	if result := ac.DB.Create(&access); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Access created successfully",
		"access":  access,
	})
}

// ListAccesses returns all accesses for a file
func (ac *AccessController) ListAccesses(c *gin.Context) {
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

	// Check if the user owns the file
	var file models.File
	if err := ac.DB.First(&file, uint(fileID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view accesses for this file"})
		return
	}

	// Fetch all accesses for the file
	var accesses []models.Access
	if err := ac.DB.Where("file_id = ?", uint(fileID)).Find(&accesses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accesses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accesses": accesses})
}

// UpdateAccess modifies an existing access record
func (ac *AccessController) UpdateAccess(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	accessID, err := strconv.ParseUint(c.Param("accessID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access ID"})
		return
	}

	// Find the access record and associated file
	var access models.Access
	if err := ac.DB.First(&access, uint(accessID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Access record not found"})
		return
	}

	// Check if the user owns the file
	var file models.File
	if err := ac.DB.First(&file, access.FileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this access"})
		return
	}

	// Parse request body
	var updateRequest struct {
		Name    string `json:"name"`
		Subnet  string `json:"subnet"`
		IP      string `json:"ip"`
		Expires string `json:"expires"`
		Public  bool   `json:"public"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update the access record
	access.Name = updateRequest.Name
	access.Subnet = updateRequest.Subnet
	access.IP = updateRequest.IP
	access.Expires = updateRequest.Expires
	access.Public = updateRequest.Public

	if err := ac.DB.Save(&access).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update access record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Access updated successfully",
		"access":  access,
	})
}

// DeleteAccess removes an access record
func (ac *AccessController) DeleteAccess(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	accessID, err := strconv.ParseUint(c.Param("accessID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access ID"})
		return
	}

	// Find the access record and associated file
	var access models.Access
	if err := ac.DB.First(&access, uint(accessID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Access record not found"})
		return
	}

	// Check if the user owns the file
	var file models.File
	if err := ac.DB.First(&file, access.FileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this access"})
		return
	}

	// Delete the access record
	if err := ac.DB.Delete(&access).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete access record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Access deleted successfully"})
}
