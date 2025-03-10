package models

import (
	"gorm.io/gorm"
)

type Access struct {
	gorm.Model
	AccessID string `gorm:"unique;not null"`
	File     File
	Link     string
	Subnet   string
	IP       string
	Expires  string
}
