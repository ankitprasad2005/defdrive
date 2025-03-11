package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Username string `gorm:"unique"` // Unique username for login identification
	Password string
	
	Files    []File `gorm:"foreignKey:UserID;references:ID"` // One-to-many relationship with File model
}
