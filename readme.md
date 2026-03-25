# Go Gin REST API

A simple Todo CRUD REST API built with [Gin](https://github.com/gin-gonic/gin), [GORM](https://gorm.io/), and PostgreSQL.

## Project Structure

```
├── main.go              # Entry point
├── config/db.go         # Database connection
├── models/todo.go       # Todo model
├── controllers/         # Route handlers (CRUD)
├── routes/routes.go     # API route definitions
├── Dockerfile           # Multi-stage Docker build
└── docker-compose.yml   # App + Postgres + Adminer
```

## Prerequisites

- Go 1.26+
- PostgreSQL 15+ (or Docker)

## Environment Variables

| Variable      | Description       | Default (Docker) |
|---------------|-------------------|------------------|
| `DB_HOST`     | Database host     | `db`             |
| `DB_USER`     | Database user     | `postgres`       |
| `DB_PASSWORD` | Database password | `postgres`       |
| `DB_NAME`     | Database name     | `todo_db`        |
| `DB_PORT`     | Database port     | `5432`           |

## Running with Docker Compose

```bash
docker-compose up --build
```

This starts three services:

| Service  | URL                    |
|----------|------------------------|
| API      | http://localhost:8080   |
| Adminer  | http://localhost:8081   |
| Postgres | localhost:5432          |

## Running Locally

Set the environment variables, then:

```bash
go run main.go
```

The server starts on port **8080**.

## API Endpoints

All routes are under `/api`.

| Method   | Endpoint         | Description       |
|----------|------------------|-------------------|
| `POST`   | `/api/todos`     | Create a todo     |
| `GET`    | `/api/todos`     | List all todos    |
| `GET`    | `/api/todos/:id` | Get a todo by ID  |
| `PUT`    | `/api/todos/:id` | Update a todo     |
| `DELETE` | `/api/todos/:id` | Delete a todo     |

### Example Request

```bash
# Create
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy groceries", "completed": false}'

# List all
curl http://localhost:8080/api/todos

# Get one
curl http://localhost:8080/api/todos/1

# Update
curl -X PUT http://localhost:8080/api/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy groceries", "completed": true}'

# Delete
curl -X DELETE http://localhost:8080/api/todos/1
```

## Todo Model

```json
{
  "ID": 1,
  "CreatedAt": "2026-03-25T00:00:00Z",
  "UpdatedAt": "2026-03-25T00:00:00Z",
  "DeletedAt": null,
  "title": "Buy groceries",
  "completed": false
}
```