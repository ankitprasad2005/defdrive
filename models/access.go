package models

import (
	"gorm.io/gorm"
)

type Access struct {
	gorm.Model
	Name       string
	Link       string   `gorm:"uniqueIndex"` // Unique index to ensure the link is unique
	Subnets    []string `gorm:"type:text[]"` // Array of subnets
	IPs        []string `gorm:"type:text[]"` // Array of IPs
	Expires    string
	Public     bool `gorm:"default:false"` // Flag indicating if access is public or restricted
	OneTimeUse bool `gorm:"default:false"` // Flag indicating if the link is one-time use
	Used       bool `gorm:"default:false"` // Flag indicating if the link has been used
	TTL        int  `gorm:"default:0"`     // Time to live (number of hops)
	EnableTTL  bool `gorm:"default:false"` // Flag to enable or disable TTL

	FileID uint `gorm:"index"`                           // Foreign key referencing the File model, indexed for query performance
	File   File `gorm:"foreignKey:FileID;references:ID"` // Relationship to File model
}
