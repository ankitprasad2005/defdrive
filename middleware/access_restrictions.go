package middleware

import (
	"defdrive/models"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AccessRestrictions middleware to handle link expiration, one-time use, subnet restriction, and public IP restriction
func AccessRestrictions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		link := c.Param("link")

		var access models.Access
		if err := db.Where("link = ?", link).First(&access).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found"})
			c.Abort()
			return
		}

		// Check expiration
		if access.Expires != "" {
			expiryTime, err := time.Parse(time.RFC3339, access.Expires)
			if err != nil || time.Now().After(expiryTime) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access link has expired"})
				c.Abort()
				return
			}
		}

		// Check one-time use
		if access.OneTimeUse && access.Used {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access link has already been used"})
			c.Abort()
			return
		}

		// Check subnet restriction
		if len(access.Subnets) > 0 {
			ip := c.ClientIP()
			allowed := false
			for _, subnet := range access.Subnets {
				_, parsedSubnet, err := net.ParseCIDR(subnet)
				if err == nil && parsedSubnet.Contains(net.ParseIP(ip)) {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access restricted to specific subnets"})
				c.Abort()
				return
			}
		}

		// Check public IP restriction
		if len(access.IPs) > 0 {
			ip := c.ClientIP()
			allowed := false
			for _, allowedIP := range access.IPs {
				if ip == allowedIP {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access restricted to specific IPs"})
				c.Abort()
				return
			}
		}

		// Mark as used if one-time use
		if access.OneTimeUse {
			access.Used = true
			db.Save(&access)
		}

		c.Next()
	}
}
