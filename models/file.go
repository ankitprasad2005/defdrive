package models

import (
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	FileID   string `gorm:"unique;not null"`
	Name     string
	Location string
	OwnerID  string `gorm:"index"` // Foreign key to User.UserID
	Owner    User   `gorm:"foreignKey:OwnerID;references:UserID"`
	Public   bool   `gorm:"default:false"` // Renamed from Access to Public for clarity, default is private (false)
	Size     int64
	Hash     string
	Accesses []Access `gorm:"foreignKey:FileID;references:FileID"`
}
