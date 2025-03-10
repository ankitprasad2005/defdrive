package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID   string `gorm:"unique;not null"`
	Name     string
	Email    string
	Username string `gorm:"unique"`
	Password string // salt saulted with default cost
}
