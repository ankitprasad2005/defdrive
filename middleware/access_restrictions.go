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
		if access.Subnet != "" {
			ip := c.ClientIP()
			_, subnet, err := net.ParseCIDR(access.Subnet)
			if err != nil || !subnet.Contains(net.ParseIP(ip)) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access restricted to specific subnet"})
				c.Abort()
				return
			}
		}

		// Check public IP restriction
		if access.IP != "" {
			ip := c.ClientIP()
			if ip != access.IP {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access restricted to specific IP"})
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
