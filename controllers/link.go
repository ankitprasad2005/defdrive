package controllers

import (
	"defdrive/models"
	// "net"
	"net/http"
	"path/filepath"
	// "time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LinkController struct {
	DB *gorm.DB
}

// NewLinkController creates a new link controller
func NewLinkController(db *gorm.DB) *LinkController {
	return &LinkController{DB: db}
}

// HandleAccessLink processes access links at /link/:hash
func (lc *LinkController) HandleAccessLink(c *gin.Context) {
	link := c.Param("hash")

	// Find the access record by link
	var access models.Access
	if err := lc.DB.Where("link = ?", link).First(&access).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found"})
		return
	}

	// Fetch the file details
	var file models.File
	if err := lc.DB.Where("id = ?", access.FileID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// THESE CHECKS ARE IMPLEMENTED IN THE access_restrictions MIDDLEWARE
	// // Check if the file is public
	// if !file.Public {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. File is not public"})
	// 	return
	// }

	// // Check if the access is public
	// if !access.Public {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Link is not public"})
	// 	return
	// }

	// // Check IP filtering
	// clientIP := c.ClientIP()
	// ipAllowed := false
	// for _, ip := range access.IPs {
	// 	if clientIP == ip {
	// 		ipAllowed = true
	// 		break
	// 	}
	// }
	// if len(access.IPs) > 0 && !ipAllowed {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. IP not allowed"})
	// 	return
	// }

	// // Check subnet filtering
	// subnetAllowed := false
	// clientIPParsed := net.ParseIP(clientIP)
	// for _, subnet := range access.Subnets {
	// 	_, cidr, err := net.ParseCIDR(subnet)
	// 	if err != nil {
	// 		continue // Skip invalid subnets
	// 	}
	// 	if cidr.Contains(clientIPParsed) {
	// 		subnetAllowed = true
	// 		break
	// 	}
	// }
	// if len(access.Subnets) > 0 && !subnetAllowed {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Subnet not allowed"})
	// 	return
	// }

	// // Check access expiration
	// if access.Expires != "" {
	// 	expiryTime, err := time.Parse(time.RFC3339, access.Expires)
	// 	if err != nil || time.Now().After(expiryTime) {
	// 		c.JSON(http.StatusForbidden, gin.H{"error": "Access link has expired"})
	// 		return
	// 	}
	// }

	// // Check one-time use restriction
	// if access.OneTimeUse && access.Used {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access link has already been used"})
	// 	return
	// }

	// // Mark as used if one-time use
	// if access.OneTimeUse {
	// 	access.Used = true
	// 	lc.DB.Save(&access)
	// }

	// Serve the file as a download using the Location field
	c.FileAttachment(filepath.Join("/app/data/uploads", filepath.Base(file.Location)), file.Name)
}
