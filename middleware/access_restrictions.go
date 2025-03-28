package middleware

import (
	"defdrive/models"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AccessRestrictions middleware to handle link expiration, one-time use, subnet restriction, public IP restriction, and TTL
func AccessRestrictions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		link := c.Param("link")
		if link == "" {
			link = c.Param("hash")
		}

		var access models.Access
		if err := db.Where("link = ?", link).First(&access).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found"})
			c.Abort()
			return
		}

		// Check if the access corresponds to a file
		var file models.File
		if err := db.Where("id = ?", access.FileID).First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			c.Abort()
			return
		}

		// Return an error if neither the file nor the access is public
		if !(file.Public && access.Public) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: neither the file nor the access is public"})
			c.Abort()
			return
		}

		// Check subnet, IP, one-time use, and TTL restrictions
		if !checkSubnetRestriction(access, c) ||
			!checkIPRestriction(access, c) ||
			!checkOneTimeUse(access, c) ||
			!checkTTL(access, db, c) ||
			!checkExpiration(access, c) {
			return
		}

		if access.OneTimeUse {
			access.Used = true
			db.Save(&access)
		}

		c.Next()
	}
}

func checkExpiration(access models.Access, c *gin.Context) bool {
	if access.Expires != "" {
		expiryTime, err := time.Parse(time.RFC3339, access.Expires)
		if err != nil || time.Now().After(expiryTime) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access link has expired"})
			c.Abort()
			return false
		}
	}
	return true
}

func checkOneTimeUse(access models.Access, c *gin.Context) bool {
	if access.OneTimeUse && access.Used {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access link has already been used"})
		c.Abort()
		return false
	}
	return true
}

func checkSubnetRestriction(access models.Access, c *gin.Context) bool {
	if len(access.Subnets) > 0 {
		ip := c.ClientIP()
		for _, subnet := range access.Subnets {
			_, parsedSubnet, err := net.ParseCIDR(subnet)
			if err == nil && parsedSubnet.Contains(net.ParseIP(ip)) {
				return true
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Access restricted to specific subnets"})
		c.Abort()
		return false
	}
	return true
}

func checkIPRestriction(access models.Access, c *gin.Context) bool {
	if len(access.IPs) > 0 {
		ip := c.ClientIP()
		for _, allowedIP := range access.IPs {
			if ip == allowedIP {
				return true
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Access restricted to specific IPs"})
		c.Abort()
		return false
	}
	return true
}

func checkTTL(access models.Access, db *gorm.DB, c *gin.Context) bool {
	if access.EnableTTL && access.TTL > 0 {
		access.TTL--
		if access.TTL == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access link has reached its TTL limit"})
			c.Abort()
			return false
		}
		db.Save(&access)
	}
	return true
}
