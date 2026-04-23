// @title           Series Tracker API
// @version         1.0
// @description     API REST para gestionar series de televisión.
// @host            series-tracker-backend.onrender.com
// @BasePath        /

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "series-tracker-backend/docs"
	"series-tracker-backend/db"
	"series-tracker-backend/handlers"
	"series-tracker-backend/middleware"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	db.Connect()

	r := mux.NewRouter()
	r.Use(middleware.CORS)

	// Swagger UI disponible en /swagger/index.html
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Servir imágenes estáticas
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
	fmt.Println("Swagger UI en http://localhost:" + port + "/swagger/index.html")
	log.Fatal(http.ListenAndServe(":"+port, r))
}