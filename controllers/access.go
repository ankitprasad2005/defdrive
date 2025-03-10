package controllers

import (
	"defdrive/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccessController struct {
	DB *gorm.DB
}

func NewAccessController(db *gorm.DB) *AccessController {
	return &AccessController{DB: db}
}

// CreateAccess generates a new access entry for a file
func (ac *AccessController) CreateAccess(c *gin.Context) {
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
	if result := ac.DB.Where("file_id = ? AND owner_id = ?", fileID, userID).First(&file); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or you don't have permission"})
		return
	}

	// Parse request body
	var input struct {
		Name    string `json:"name"` // Added name field
		Subnet  string `json:"subnet"`
		IP      string `json:"ip"`
		Expires string `json:"expires"`
		Public  bool   `json:"public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate expiration date if provided
	if input.Expires != "" {
		_, err := time.Parse(time.RFC3339, input.Expires)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expiration date format. Use RFC3339 format (e.g. 2023-01-01T15:04:05Z)"})
			return
		}
	}

	// Create access record
	access := models.Access{
		AccessID: uuid.New().String(),
		FileID:   fileID,
		Name:     input.Name,          // Set name from input
		Link:     uuid.New().String(), // Generate a random link ID
		Subnet:   input.Subnet,
		IP:       input.IP,
		Expires:  input.Expires,
		Public:   input.Public,
	}

	if result := ac.DB.Create(&access); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Access created successfully",
		"access":  access,
	})
}

// ListAccesses retrieves all accesses for a specific file
func (ac *AccessController) ListAccesses(c *gin.Context) {
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

	// Check if user owns the file
	var file models.File
	if result := ac.DB.Where("file_id = ? AND owner_id = ?", fileID, userID).First(&file); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found or you don't have permission"})
		return
	}

	// Get all accesses for the file
	var accesses []models.Access
	if result := ac.DB.Where("file_id = ?", fileID).Find(&accesses); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accesses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accesses": accesses,
	})
}

// UpdateAccess allows modifying an existing access entry
func (ac *AccessController) UpdateAccess(c *gin.Context) {
	// Get the logged-in user from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get access ID from URL parameter
	accessID := c.Param("accessID")
	if accessID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access ID is required"})
		return
	}

	// Find the access entry
	var access models.Access
	if result := ac.DB.Where("access_id = ?", accessID).First(&access); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Access entry not found"})
		return
	}

	// Verify the user owns the file associated with this access
	var file models.File
	if result := ac.DB.Where("file_id = ? AND owner_id = ?", access.FileID, userID).First(&file); result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this access"})
		return
	}

	// Parse request body
	var input struct {
		Name    string `json:"name"` // Added name field
		Subnet  string `json:"subnet"`
		IP      string `json:"ip"`
		Expires string `json:"expires"`
		Public  bool   `json:"public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate expiration date if provided
	if input.Expires != "" {
		_, err := time.Parse(time.RFC3339, input.Expires)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expiration date format. Use RFC3339 format (e.g. 2023-01-01T15:04:05Z)"})
			return
		}
	}

	// Update access fields
	access.Name = input.Name // Update name
	access.Subnet = input.Subnet
	access.IP = input.IP
	access.Expires = input.Expires
	access.Public = input.Public

	// Save the updated access entry
	if result := ac.DB.Save(&access); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Access updated successfully",
		"access":  access,
	})
}

// DeleteAccess removes an existing access entry
func (ac *AccessController) DeleteAccess(c *gin.Context) {
	// Get the logged-in user from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get access ID from URL parameter
	accessID := c.Param("accessID")
	if accessID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access ID is required"})
		return
	}

	// Find the access entry
	var access models.Access
	if result := ac.DB.Where("access_id = ?", accessID).First(&access); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Access entry not found"})
		return
	}

	// Verify the user owns the file associated with this access
	var file models.File
	if result := ac.DB.Where("file_id = ? AND owner_id = ?", access.FileID, userID).First(&file); result.Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this access"})
		return
	}

	// Delete the access entry
	if result := ac.DB.Delete(&access); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Access deleted successfully",
	})
}
