package controllers

import (
	"defdrive/models"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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
	go func() {
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
	}()
}

// Login authenticates a user
func (uc *UserController) Login(c *gin.Context) {
	go func() {
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
	}()
}
