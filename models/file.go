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
	Public   bool `gorm:"default:false"`
	
	UserID   uint `gorm:"index"`
	User     User `gorm:"foreignKey:UserID;references:ID"`
	
	Accesses []Access `gorm:"foreignKey:FileID;references:ID"` // One-to-many relationship with Access model
}
