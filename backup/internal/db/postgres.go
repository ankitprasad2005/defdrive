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
        User:     os.Getenv("POSTGRES_USER"),
        Password: os.Getenv("POSTGRES_PASSWORD"),
        Database: os.Getenv("POSTGRES_DB"),
    })

	log.Println("DB: %v", DB)

    if err := DB.Ping(context.Background()); err != nil {
		log.Println("DB: %v", DB)
        log.Fatal(err)
    }
}