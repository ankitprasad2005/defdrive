package routes

import (
    "defdrive/internal/handlers"

    "github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
    router := mux.NewRouter()

    // User routes
    router.HandleFunc("/signin", handlers.Signin).Methods("POST")
    router.HandleFunc("/login", handlers.Login).Methods("POST")
    router.HandleFunc("/user", handlers.CreateUser).Methods("POST")
    router.HandleFunc("/user", handlers.GetUser).Methods("GET")

    // File routes
    router.HandleFunc("/file", handlers.CreateFile).Methods("POST")
    router.HandleFunc("/file/{fileID}", handlers.GetFile).Methods("GET")

    // Access routes
    router.HandleFunc("/access", handlers.CreateAccess).Methods("POST")
    router.HandleFunc("/access/{accessID}", handlers.GetAccess).Methods("GET")

    return router
}