package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Use test database configuration
	dbHost := getEnv("TEST_DB_HOST", "localhost")
	dbPort := getEnv("TEST_DB_PORT", "5432")
	dbUser := getEnv("TEST_DB_USER", "postgres")
	dbPass := getEnv("TEST_DB_PASSWORD", "postgres")
	dbName := getEnv("TEST_DB_NAME", "logmeup_test")

	// Connect to the test database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Drop all tables
	_, err := db.Exec(`
		DROP TABLE IF EXISTS actions;
		DROP TABLE IF EXISTS notes;
	`)
	if err != nil {
		t.Fatalf("Failed to clean up test database: %v", err)
	}

	// Close the database connection
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close test database connection: %v", err)
	}
}

// SetupTestSchema creates the test database schema
func SetupTestSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	// Create tables
	_, err := db.Exec(`
		CREATE TABLE notes (
			id BIGSERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			date DATE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE TABLE actions (
			id BIGSERIAL PRIMARY KEY,
			note_id BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
			description TEXT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE INDEX idx_notes_date ON notes(date);
		CREATE INDEX idx_actions_note_id ON actions(note_id);
	`)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
