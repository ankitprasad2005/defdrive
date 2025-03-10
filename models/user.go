package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID   uint `gorm:"primaryKey"` // Unique identifier for each user, serves as primary key
	Name     string
	Email    string
	Username string `gorm:"unique"` // Unique username for login identification
	Password string
}
