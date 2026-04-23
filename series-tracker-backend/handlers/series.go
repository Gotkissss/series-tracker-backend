package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"series-tracker-backend/db"
	"series-tracker-backend/models"

	"github.com/gorilla/mux"
)

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// GetAllSeries godoc
// @Summary      Listar series
// @Description  Retorna todas las series. Soporta búsqueda, paginación y ordenamiento.
// @Tags         series
// @Produce      json
// @Param        q      query  string  false  "Buscar por nombre"
// @Param        page   query  int     false  "Número de página (default: 1)"
// @Param        limit  query  int     false  "Resultados por página (default: 10)"
// @Param        sort   query  string  false  "Campo por el que ordenar"
// @Param        order  query  string  false  "asc o desc"
// @Success      200  {array}   models.Series
// @Failure      500  {object}  map[string]string
// @Router       /series [get]
func GetAllSeries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	page := 1
	limit := 10
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 { page = p }
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 { limit = l }
	offset := (page - 1) * limit

	validSort := map[string]bool{"title": true, "genre": true, "status": true, "episodes": true, "rating": true, "created_at": true}
	if !validSort[sort] { sort = "created_at" }
	if order != "asc" && order != "desc" { order = "desc" }

	query := `SELECT id, title, genre, status, episodes, rating, COALESCE(image_data, ''), created_at FROM series WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if q != "" {
		query += ` AND title ILIKE $` + strconv.Itoa(argIdx)
		args = append(args, "%"+q+"%")
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d OFFSET $%d`, sort, order, argIdx, argIdx+1)
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
		if err := rows.Scan(&s.ID, &s.Title, &s.Genre, &s.Status, &s.Episodes, &s.Rating, &s.ImageURL, &s.CreatedAt); err != nil {
			jsonError(w, http.StatusInternalServerError, "Error leyendo datos")
			return
		}
		series = append(series, s)
	}
	jsonResponse(w, http.StatusOK, series)
}

// GetSeriesByID godoc
// @Summary      Obtener una serie
// @Description  Retorna una serie por su ID
// @Tags         series
// @Produce      json
// @Param        id  path  int  true  "ID de la serie"
// @Success      200  {object}  models.Series
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /series/{id} [get]
func GetSeriesByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var s models.Series
	err = db.DB.QueryRow(
		`SELECT id, title, genre, status, episodes, rating, COALESCE(image_data, ''), created_at FROM series WHERE id = $1`, id,
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

// CreateSeries godoc
// @Summary      Crear una serie
// @Description  Crea una nueva serie en la base de datos
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        series  body  models.Series  true  "Datos de la serie"
// @Success      201  {object}  models.Series
// @Failure      400  {object}  map[string]string
// @Router       /series [post]
func CreateSeries(w http.ResponseWriter, r *http.Request) {
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

	err := db.DB.QueryRow(
		`INSERT INTO series (title, genre, status, episodes, rating) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		s.Title, s.Genre, s.Status, s.Episodes, s.Rating,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error creando la serie")
		return
	}
	jsonResponse(w, http.StatusCreated, s)
}

// UpdateSeries godoc
// @Summary      Editar una serie
// @Description  Actualiza los datos de una serie existente
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        id      path  int           true  "ID de la serie"
// @Param        series  body  models.Series  true  "Datos actualizados"
// @Success      200  {object}  models.Series
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /series/{id} [put]
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
		`UPDATE series SET title=$1, genre=$2, status=$3, episodes=$4, rating=$5 WHERE id=$6`,
		s.Title, s.Genre, s.Status, s.Episodes, s.Rating, id,
	)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error actualizando la serie")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}
	s.ID = id
	jsonResponse(w, http.StatusOK, s)
}

// DeleteSeries godoc
// @Summary      Eliminar una serie
// @Description  Elimina una serie por su ID
// @Tags         series
// @Param        id  path  int  true  "ID de la serie"
// @Success      204
// @Failure      404  {object}  map[string]string
// @Router       /series/{id} [delete]
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
	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadImage godoc
// @Summary      Subir imagen de una serie
// @Description  Sube una imagen (JPG/PNG/WEBP, max 1MB) y la guarda en la base de datos
// @Tags         series
// @Accept       multipart/form-data
// @Produce      json
// @Param        id     path      int   true  "ID de la serie"
// @Param        image  formData  file  true  "Archivo de imagen"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /series/{id}/image [post]
func UploadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		jsonError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var exists bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM series WHERE id = $1)`, id).Scan(&exists)
	if !exists {
		jsonError(w, http.StatusNotFound, "Serie no encontrada")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		jsonError(w, http.StatusBadRequest, "La imagen no puede superar 1MB")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		jsonError(w, http.StatusBadRequest, "No se encontró el campo 'image'")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		jsonError(w, http.StatusBadRequest, "Solo se permiten JPG, PNG o WEBP")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error leyendo imagen")
		return
	}

	b64 := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)

	_, err = db.DB.Exec(`UPDATE series SET image_data = $1 WHERE id = $2`, b64, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Error guardando imagen