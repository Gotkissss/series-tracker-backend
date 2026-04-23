package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Fallback para desarrollo local
		connStr = "host=localhost port=5432 user=postgres password=postgres dbname=seriesdb sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error abriendo conexión a la base de datos:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Error conectando a la base de datos:", err)
	}

	fmt.Println("Conectado a PostgreSQL")
}