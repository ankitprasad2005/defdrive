package controllers

import (
	"defdrive/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

// NewUserController creates a new user controller
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// SignUp handles user registration
func (uc *UserController) SignUp(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Create the user in the database
	result := uc.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + result.Error.Error()})
		return
	}

	// Don't return the password in the response
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "user": user})
}

// Login authenticates a user
func (uc *UserController) Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := uc.DB.Where("username = ?", loginRequest.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Compare the provided password with the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"name":     user.Name,
		},
	})
}

// GetUserLimits returns the current user's limits and usage
func (uc *UserController) GetUserLimits(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := uc.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Get current file count
	var currentFileCount int64
	if err := uc.DB.Model(&models.File{}).Where("user_id = ?", userID).Count(&currentFileCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current file count"})
		return
	}

	// Get current storage usage
	var currentStorage int64
	if err := uc.DB.Model(&models.File{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&currentStorage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current storage usage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"max_files":         user.MaxFiles,
		"max_storage":       user.MaxStorage,
		"current_files":     currentFileCount,
		"current_storage":   currentStorage,
		"remaining_files":   user.MaxFiles - int(currentFileCount),
		"remaining_storage": user.MaxStorage - currentStorage,
	})
}

// UpdateUserLimits allows updating user limits (admin only for now)
func (uc *UserController) UpdateUserLimits(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var updateRequest struct {
		MaxFiles   *int   `json:"max_files"`
		MaxStorage *int64 `json:"max_storage"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := uc.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update limits if provided
	if updateRequest.MaxFiles != nil {
		if *updateRequest.MaxFiles < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Max files cannot be negative"})
			return
		}
		user.MaxFiles = *updateRequest.MaxFiles
	}

	if updateRequest.MaxStorage != nil {
		if *updateRequest.MaxStorage < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Max storage cannot be negative"})
			return
		}
		user.MaxStorage = *updateRequest.MaxStorage
	}

	if err := uc.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user limits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User limits updated successfully",
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"max_files":   user.MaxFiles,
			"max_storage": user.MaxStorage,
		},
	})
}

// GetAllUsersLimits returns limits and usage for all users (admin endpoint)
func (uc *UserController) GetAllUsersLimits(c *gin.Context) {
	var users []models.User
	if err := uc.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	var userLimits []gin.H
	for _, user := range users {
		// Get current file count for each user
		var currentFileCount int64
		uc.DB.Model(&models.File{}).Where("user_id = ?", user.ID).Count(&currentFileCount)

		// Get current storage usage for each user
		var currentStorage int64
		uc.DB.Model(&models.File{}).Where("user_id = ?", user.ID).Select("COALESCE(SUM(size), 0)").Scan(&currentStorage)

		userLimits = append(userLimits, gin.H{
			"user_id":           user.ID,
			"username":          user.Username,
			"email":             user.Email,
			"max_files":         user.MaxFiles,
			"max_storage":       user.MaxStorage,
			"current_files":     currentFileCount,
			"current_storage":   currentStorage,
			"remaining_files":   user.MaxFiles - int(currentFileCount),
			"remaining_storage": user.MaxStorage - currentStorage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userLimits,
	})
}
