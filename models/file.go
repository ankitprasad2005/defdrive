package models

import (
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	Name     string
	Location string
	Size     int64
	Hash     string
	Public   bool `gorm:"default:false"`                   // Flag indicating if file is publicly accessible
	
	UserID   uint `gorm:"index"`                           // Foreign key linking to the User who owns this file, indexed for performance
	User     User `gorm:"foreignKey:UserID;references:ID"` // Relationship to User model
	
	Accesses []Access `gorm:"foreignKey:FileID;references:ID"` // One-to-many relationship with Access model
}
