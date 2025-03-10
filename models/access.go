package models

import (
	"gorm.io/gorm"
)

type Access struct {
	gorm.Model
	AccessID string `gorm:"unique;not null"`
	FileID   string `gorm:"index"` // Foreign key to File.FileID
	Name     string // Added name field for better identification
	Link     string
	Subnet   string
	IP       string
	Expires  string
	Public   bool `gorm:"default:false"` // New field for public/private access control
}
