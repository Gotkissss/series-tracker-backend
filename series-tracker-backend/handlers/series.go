package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"series-tracker-backend/db"
	"series-tracker-backend/models"

	"github.com/gorilla/mux"
)

// responde con JSON y un código HTTP
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// responde con un error en JSON
func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// GET /series
func GetAllSeries(w http.ResponseWriter, r *http.Request) {
	// Parámetros de búsqueda, paginación y ordenamiento
	q := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	// Valores por defecto
	page := 1
	limit := 10
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := (page - 1) * limit

	// Columnas válidas para ordenar (evita SQL injection)
	validSortColumns := map[string]bool{
		"title": true, "genre": true, "status": true,
		"episodes": true, "rating": true, "created_at": true,
	}
	if !validSortColumns[sort] {
		sort = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Query base
	query := `SELECT id, title, genre, status, episodes, rating, image_url, created_at
	          FROM series WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if q != "" {
		query += ` AND title ILIKE $` + strconv.Itoa(argIdx)
		args = append(args, "%"+q+"%")
		argIdx++
	}

	query += ` ORDER BY ` + sort + ` ` + order
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error consultando la base de datos")
		return
	}
	defer rows.Close()

	series := []models.Series{}
	for rows.Next() {
		var s models.Series
		err := rows.Scan(&s.ID, &s.Title, &s.Genre, &s.Status, &s.Episodes, &s.Rating, &s.ImageURL, &s.CreatedAt)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Error leyendo datos")
			return
		}
		series = append(series, s)
	}

	jsonResponse(w, http.StatusOK, series)
}

// GET /series/:id
func GetSeriesByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var s models.Series
	err = db.DB.QueryRow(
		`SELECT id, title, genre, status, episodes, rating, image_url, created_at FROM series WHERE id = $1`, id,
	).Scan(&s.ID, &s.Title, &s.Genre, &s.Status, &s.Episodes, &s.Rating, &s.ImageURL, &s.CreatedAt)

	if err == sql.ErrNoRows {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error consultando la base de datos")
		return
	}

	jsonResponse(w, http.StatusOK, s)
}

// POST /series
func CreateSeries(w http.ResponseWriter, r *http.Request) {
	var s models.Series
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		jsonError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// Validación
	if s.Title == "" {
		jsonError(w, http.StatusBadRequest, "El campo 'title' es obligatorio")
		return
	}
	if s.Episodes < 0 {
		jsonError(w, http.StatusBadRequest, "El campo 'episodes' no puede ser negativo")
		return
	}
	if s.Rating < 0 || s.Rating > 10 {
		jsonError(w, http.StatusBadRequest, "El campo 'rating' debe estar entre 0 y 10")
		return
	}

	err := db.DB.QueryRow(
		`INSERT INTO series (title, genre, status, episodes, rating, image_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		s.Title, s.Genre, s.Status, s.Episodes, s.Rating, s.ImageURL,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error creando la serie")
		return
	}

	jsonResponse(w, http.StatusCreated, s)
}

// PUT /series/:id
func UpdateSeries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var s models.Series
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		jsonError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if s.Title == "" {
		jsonError(w, http.StatusBadRequest, "El campo 'title' es obligatorio")
		return
	}
	if s.Episodes < 0 {
		jsonError(w, http.StatusBadRequest, "El campo 'episodes' no puede ser negativo")
		return
	}
	if s.Rating < 0 || s.Rating > 10 {
		jsonError(w, http.StatusBadRequest, "El campo 'rating' debe estar entre 0 y 10")
		return
	}

	result, err := db.DB.Exec(
		`UPDATE series SET title=$1, genre=$2, status=$3, episodes=$4, rating=$5, image_url=$6 WHERE id=$7`,
		s.Title, s.Genre, s.Status, s.Episodes, s.Rating, s.ImageURL, id,
	)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error actualizando la serie")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}

	s.ID = id
	jsonResponse(w, http.StatusOK, s)
}

// DELETE /series/:id
func DeleteSeries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	result, err := db.DB.Exec(`DELETE FROM series WHERE id = $1`, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error eliminando la serie")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 - sin body
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// Verificar que la serie existe
	var exists bool
	err = db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM series WHERE id = $1)`, id).Scan(&exists)
	if err != nil || !exists {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}

	// Limitar tamaño a 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		jsonError(w, http.StatusBadRequest, "La imagen no puede superar 1MB")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		jsonError(w, http.StatusBadRequest, "No se encontró el campo 'image' en el formulario")
		return
	}
	defer file.Close()

	// Validar tipo de archivo
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		jsonError(w, http.StatusBadRequest, "Solo se permiten imágenes JPG, PNG o WEBP")
		return
	}

	// Guardar el archivo con nombre único basado en el ID
	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	} else if contentType == "image/webp" {
		ext = ".webp"
	}

	filename := "series_" + strconv.Itoa(id) + ext
	filepath := "uploads/" + filename

	dst, err := os.Create(filepath)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error guardando la imagen")
		return
	}
	defer dst.Close()

	buf := make([]byte, 1<<20)
	n, _ := file.Read(buf)
	dst.Write(buf[:n])

	// Construir la URL pública
	scheme := "http"
	imageURL := scheme + "://" + r.Host + "/uploads/" + filename

	// Actualizar image_url en la base de datos
	_, err = db.DB.Exec(`UPDATE series SET image_url = $1 WHERE id = $2`, imageURL, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error actualizando la URL de imagen")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"image_url": imageURL})
}
