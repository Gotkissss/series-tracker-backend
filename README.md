# Series Tracker — Backend

API REST para gestionar series de televisión, construida con Go y PostgreSQL.

## Stack
- Go 1.22 con gorilla/mux
- PostgreSQL (Neon.tech)
- Deploy en Render

## Links
- Backend en producción: https://series-tracker-backend-rymf.onrender.com
- Frontend: https://bespoke-sherbet-009345.netlify.app
- Swagger UI: https://series-tracker-backend-rymf.onrender.com/swagger/index.html

## Correr localmente

1. Tener Go 1.22+ y PostgreSQL instalados
2. Crear base de datos `seriesdb` en PostgreSQL local
3. Ejecutar el SQL de creación de tabla (ver abajo)
4. Clonar el repositorio
5. Desde la carpeta del proyecto:

```bash
go run main.go
```

El servidor corre en `http://localhost:8080`

## SQL de la tabla

```sql
CREATE TABLE series (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    genre VARCHAR(100),
    status VARCHAR(50),
    episodes INT,
    rating DECIMAL(3,1),
    image_url TEXT,
    image_data TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | /series | Listar series (paginación, búsqueda, ordenamiento) |
| GET | /series/:id | Obtener una serie |
| POST | /series | Crear serie |
| PUT | /series/:id | Editar serie |
| DELETE | /series/:id | Eliminar serie |
| POST | /series/:id/image | Subir imagen |

## CORS

CORS es una política de seguridad del navegador que bloquea peticiones HTTP entre orígenes distintos (distinto dominio o puerto). Se configuró con `Access-Control-Allow-Origin: *` para permitir peticiones desde cualquier origen durante desarrollo y producción.

## Challenges implementados

- Swagger UI con spec OpenAPI completo
- Códigos HTTP correctos (201, 204, 404, 400)
- Validación server-side con errores en JSON
- Paginación con page y limit
- Búsqueda por nombre con q
- Ordenamiento con sort y order
- Subida de imágenes (max 1MB, guardadas como base64 en PostgreSQL)



## Screenshot
