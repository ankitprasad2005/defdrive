package handlers

import (
    "defdrive/internal/db"
    "defdrive/internal/models"
    // "github.com/go-pg/pg/v10"
)

func createFile(file *models.File) error {
    _, err := db.DB.Model(file).Insert()
    return err
}

func getFileByID(fileID string) (*models.File, error) {
    file := &models.File{}
    err := db.DB.Model(file).Where("file_id = ?", fileID).Select()
    return file, err
}