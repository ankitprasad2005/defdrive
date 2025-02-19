package routes

import (
	"defdrive/internal/handlers"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/signin", handlers.Signin).Methods("POST")
	router.HandleFunc("/login", handlers.Login).Methods("POST")
	return router
}
