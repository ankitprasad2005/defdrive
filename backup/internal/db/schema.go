package db

import (
    "log"

    "github.com/go-pg/pg/v10/orm"
    "defdrive/internal/models"
)

func CreateSchema() {
    models := []interface{}{
        (*models.User)(nil),
        (*models.File)(nil),
        (*models.Access)(nil),
    }

    for _, model := range models {
        err := DB.Model(model).CreateTable(&orm.CreateTableOptions{
            IfNotExists: true,
        })
        if err != nil {
            log.Fatalf("Error creating schema: %v", err)
        }
    }
}