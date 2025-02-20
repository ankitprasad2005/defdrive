package handlers

import (
    "defdrive/internal/db"
    "defdrive/internal/models"
    // "github.com/go-pg/pg/v10"
)

func createAccess(access *models.Access) error {
    _, err := db.DB.Model(access).Insert()
    return err
}

func getAccessByID(accessID string) (*models.Access, error) {
    access := &models.Access{}
    err := db.DB.Model(access).Where("access_id = ?", accessID).Select()
    return access, err
}