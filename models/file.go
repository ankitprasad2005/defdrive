package models

import (
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	FileID   string `gorm:"unique;not null"`
	Name     string
	Location string
	Owner    User
	Access   string
	Size     int64
	Hash     string
	Accesses []Access // Array of access entries for this file
}
