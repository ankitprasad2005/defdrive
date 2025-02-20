package main

import (
    "log"
    "net/http"
    "defdrive/config"
    "defdrive/internal/db"
    "defdrive/internal/routes"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    config.LoadConfig()
    db.Init()
    db.CreateSchema()
    router := routes.SetupRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
}