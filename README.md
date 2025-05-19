# LogMeUp API

This is the backend API for the LogMeUp application, built with Go and PostgreSQL.

## Prerequisites

- Go 1.21 or later
- PostgreSQL 14 or later
- golang-migrate CLI tool

## Setup

1. Create a PostgreSQL database:
```sql
CREATE DATABASE logmeup;
```

2. Copy the `.env.example` file to `.env` and update the values:
```bash
cp .env.example .env
```

3. Run database migrations:
```bash
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/logmeup?sslmode=disable" up
```

4. Install dependencies:
```bash
go mod tidy
```

5. Run the application:
```bash
go run cmd/api/main.go
```

## API Endpoints

### Notes

- `POST /api/notes` - Create a new note
- `GET /api/notes/:id` - Get a note by ID
- `GET /api/notes?date=YYYY-MM-DD` - Get notes by date
- `PUT /api/notes/:id` - Update a note
- `DELETE /api/notes/:id` - Delete a note

### Actions

- `POST /api/actions` - Create a new action
- `GET /api/actions/:id` - Get an action by ID
- `GET /api/actions/note/:note_id` - Get actions by note ID
- `PUT /api/actions/:id` - Update an action
- `DELETE /api/actions/:id` - Delete an action

## Development

To run the application in development mode with hot reload:

```bash
go install github.com/cosmtrek/air@latest
air
```

## Testing

Run the tests:

```bash
go test ./...
``` 