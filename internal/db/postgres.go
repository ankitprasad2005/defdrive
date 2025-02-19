package db

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-pg/pg/v10"
)

var DB *pg.DB

func Init() {
    DB = pg.Connect(&pg.Options{
        Addr:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        Database: os.Getenv("DB_NAME"),
    })

    if err := DB.Ping(context.Background()); err != nil {
        log.Fatal(err)
    }
}