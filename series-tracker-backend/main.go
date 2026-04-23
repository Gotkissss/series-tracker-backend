package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"series-tracker-backend/db"
	"series-tracker-backend/handlers"
	"series-tracker-backend/middleware"

	"github.com/gorilla/mux"
)

func main() {
	// Conectar a la base de datos
	db.Connect()

	// Crear el router
	r := mux.NewRouter()

	// Aplicar middleware de CORS a todas las rutas
	r.Use(middleware.CORS)

	// Rutas
	r.HandleFunc("/series", handlers.GetAllSeries).Methods("GET", "OPTIONS")
	r.HandleFunc("/series/{id}", handlers.GetSeriesByID).Methods("GET", "OPTIONS")
	r.HandleFunc("/series", handlers.CreateSeries).Methods("POST", "OPTIONS")
	r.HandleFunc("/series/{id}", handlers.UpdateSeries).Methods("PUT", "OPTIONS")
	r.HandleFunc("/series/{id}", handlers.DeleteSeries).Methods("DELETE", "OPTIONS")

	// Puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("🚀 Servidor corriendo en http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}