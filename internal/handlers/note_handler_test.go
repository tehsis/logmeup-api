package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tehsis/logmeup-api/internal/models"
	"github.com/tehsis/logmeup-api/internal/repository"
	"github.com/tehsis/logmeup-api/internal/testutil"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *repository.NoteRepository) {
	gin.SetMode(gin.TestMode)
	db := testutil.SetupTestDB(t)
	testutil.SetupTestSchema(t, db)

	noteRepo := repository.NewNoteRepository(db)
	noteHandler := NewNoteHandler(noteRepo)

	r := gin.Default()
	r.POST("/api/notes", noteHandler.Create)
	r.GET("/api/notes/:id", noteHandler.GetByID)
	r.GET("/api/notes", noteHandler.GetByDate)
	r.PUT("/api/notes/:id", noteHandler.Update)
	r.DELETE("/api/notes/:id", noteHandler.Delete)

	return r, noteRepo
}

func TestNoteHandler(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		r, _ := setupTestRouter(t)

		note := models.CreateNoteRequest{
			Content: "Test note",
			Date:    time.Now(),
		}
		body, _ := json.Marshal(note)

		req := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Note
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Content != note.Content {
			t.Errorf("Expected content %q, got %q", note.Content, response.Content)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		r, repo := setupTestRouter(t)

		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for GetByID",
			Date:    time.Now(),
		}
		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/notes/"+string(created.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Note
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, response.ID)
		}
	})

	t.Run("GetByDate", func(t *testing.T) {
		r, repo := setupTestRouter(t)

		// Create test notes
		date := time.Now()
		notes := []*models.CreateNoteRequest{
			{Content: "Note 1", Date: date},
			{Content: "Note 2", Date: date},
		}

		for _, note := range notes {
			_, err := repo.Create(note)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/api/notes?date="+date.Format("2006-01-02"), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response []models.Note
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(response))
		}
	})

	t.Run("Update", func(t *testing.T) {
		r, repo := setupTestRouter(t)

		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for Update",
			Date:    time.Now(),
		}
		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		update := models.UpdateNoteRequest{
			Content: "Updated content",
		}
		body, _ := json.Marshal(update)

		req := httptest.NewRequest(http.MethodPut, "/api/notes/"+string(created.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Note
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Content != update.Content {
			t.Errorf("Expected content %q, got %q", update.Content, response.Content)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		r, repo := setupTestRouter(t)

		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for Delete",
			Date:    time.Now(),
		}
		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/notes/"+string(created.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
		}

		// Verify note is deleted
		_, err = repo.GetByID(created.ID)
		if err == nil {
			t.Error("Expected error when getting deleted note")
		}
	})
}
