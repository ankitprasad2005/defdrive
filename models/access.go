package models

import (
	"gorm.io/gorm"
)

type Access struct {
	gorm.Model
	AccessID uint `gorm:"primaryKey"`                          // Unique identifier for each access record
	FileID   uint `gorm:"index"`                               // Foreign key referencing the File model, indexed for query performance
	File     File `gorm:"foreignKey:FileID;references:FileID"` // Relationship to File model
	Name     string
	Link     string
	Subnet   string
	IP       string
	Expires  string
	Public   bool `gorm:"default:false"` // Flag indicating if access is public or restricted
}
