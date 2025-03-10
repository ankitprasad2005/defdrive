package models

import (
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	FileID   uint `gorm:"primaryKey"` // Unique identifier for each file in the system
	Name     string
	Location string
	OwnerID  uint `gorm:"index"`                            // Foreign key linking to the User who owns this file, indexed for performance
	Owner    User `gorm:"foreignKey:OwnerID;references:ID"` // Relationship to User model
	Public   bool `gorm:"default:false"`                    // Flag indicating if file is publicly accessible
	Size     int64
	Hash     string
	Accesses []Access `gorm:"foreignKey:FileID;references:FileID"` // One-to-many relationship with Access model
}
