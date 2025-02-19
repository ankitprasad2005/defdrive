package main

import (
    "log"
    "net/http"
    "defdrive/internal/routes"
    "defdrive/config"
    "defdrive/internal/db"
)

func main() {
    config.LoadConfig()
    db.Init()
    router := routes.SetupRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
}