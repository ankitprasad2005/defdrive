package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Email    string
	Username string `gorm:"unique"`
	Password string
	
	MaxFiles   int   `gorm:"default:100"`        // default 100 files
	MaxStorage int64 `gorm:"default:1073741824"` // default 1GB

	Files []File `gorm:"foreignKey:UserID;references:ID"` // One-to-many relationship with File model
}
