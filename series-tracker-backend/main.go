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
	db.Connect()

	r := mux.NewRouter()
	r.Use(middleware.CORS)

	// Servir imágenes estáticas desde /uploads
	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))),
	)

	// Rutas de la API
	r.HandleFunc("/series", handlers.GetAllSeries).Methods("GET", "OPTIONS")
	r.HandleFunc("/series/{id}", handlers.GetSeriesByID).Methods("GET", "OPTIONS")
	r.HandleFunc("/series", handlers.CreateSeries).Methods("POST", "OPTIONS")
	r.HandleFunc("/series/{id}", handlers.UpdateSeries).Methods("PUT", "OPTIONS")
	r.HandleFunc("/series/{id}", handlers.DeleteSeries).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/series/{id}/image", handlers.UploadImage).Methods("POST", "OPTIONS")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Servidor corriendo en http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}