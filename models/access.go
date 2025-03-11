package models

import (
	"gorm.io/gorm"
)

type Access struct {
	gorm.Model
	Name    string
	Link    string
	Subnet  string
	IP      string
	Expires string
	Public  bool `gorm:"default:false"` // Flag indicating if access is public or restricted
	
	FileID  uint `gorm:"index"`                           // Foreign key referencing the File model, indexed for query performance
	File    File `gorm:"foreignKey:FileID;references:ID"` // Relationship to File model
}
