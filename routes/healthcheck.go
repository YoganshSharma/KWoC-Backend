package routes

import (
	"github.com/gorilla/mux"
	"kwoc20-backend/controllers"
)

func RegisterHealthCheck(r *mux.Router) {
	//Ping can be removed
	r.HandleFunc("/ping", controllers.Ping).Methods("GET")
	r.HandleFunc("", controllers.HealthCheck).Methods("GET")
}
