package controllers

import (
	"crypto/md5"
	"defdrive/models"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"time"

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

// generateRandomLink generates a random link that looks like an MD5 hash
func (ac *AccessController) generateRandomLink() string {
	for {
		hash := md5.New()
		hash.Write([]byte(uuid.New().String() + time.Now().String()))
		link := hex.EncodeToString(hash.Sum(nil)) // MD5 generates a 32-character hash

		// Check if the link already exists in the database
		var access models.Access
		if err := ac.DB.Where("link = ?", link).First(&access).Error; err == gorm.ErrRecordNotFound {
			return link
		}
	}
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
		Name       string   `json:"name"`
		Subnets    []string `json:"subnets"`
		IPs        []string `json:"ips"`
		Expires    string   `json:"expires"`
		Public     bool     `json:"public"`
		OneTimeUse bool     `json:"oneTimeUse"`
		TTL        int      `json:"ttl"`
		EnableTTL  bool     `json:"enableTTL"`
	}

	if err := c.ShouldBindJSON(&accessRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Generate a unique access link
	link := ac.generateRandomLink()

	// Create the access record
	access := models.Access{
		FileID:     uint(fileID),
		Name:       accessRequest.Name,
		Link:       link,
		Subnets:    accessRequest.Subnets,
		IPs:        accessRequest.IPs,
		Expires:    accessRequest.Expires,
		Public:     accessRequest.Public,
		OneTimeUse: accessRequest.OneTimeUse,
		TTL:        accessRequest.TTL,
		EnableTTL:  accessRequest.EnableTTL,
	}

	if result := ac.DB.Create(&access); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access record"})
		return
	}

	hostURL := os.Getenv("HOST_URL")

	c.JSON(http.StatusOK, gin.H{
		"message": "Access created successfully",
		"access":  access,
		"link":    hostURL + "/" + link,
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

// GetAccess retrieves the details of a specific access record
func (ac *AccessController) GetAccess(c *gin.Context) {
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

	var file models.File
	if err := ac.DB.First(&file, access.FileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check if the user owns the file
	if file.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access": access})
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
		Name       string   `json:"name"`
		Subnets    []string `json:"subnets"`
		IPs        []string `json:"ips"`
		Expires    string   `json:"expires"`
		Public     bool     `json:"public"`
		OneTimeUse bool     `json:"oneTimeUse"`
		TTL        int      `json:"ttl"`
		EnableTTL  bool     `json:"enableTTL"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update the access record
	access.Name = updateRequest.Name
	access.Subnets = updateRequest.Subnets
	access.IPs = updateRequest.IPs
	access.Expires = updateRequest.Expires
	access.Public = updateRequest.Public
	access.OneTimeUse = updateRequest.OneTimeUse
	access.TTL = updateRequest.TTL
	access.EnableTTL = updateRequest.EnableTTL

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
