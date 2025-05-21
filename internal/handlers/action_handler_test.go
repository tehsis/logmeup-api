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

func setupActionTestRouter(t *testing.T) (*gin.Engine, *repository.ActionRepository, *repository.NoteRepository) {
	gin.SetMode(gin.TestMode)
	db := testutil.SetupTestDB(t)
	testutil.SetupTestSchema(t, db)

	noteRepo := repository.NewNoteRepository(db)
	actionRepo := repository.NewActionRepository(db)
	actionHandler := NewActionHandler(actionRepo)

	r := gin.Default()
	r.POST("/api/actions", actionHandler.Create)
	r.GET("/api/actions/:id", actionHandler.GetByID)
	r.GET("/api/actions/note/:note_id", actionHandler.GetByNoteID)
	r.PUT("/api/actions/:id", actionHandler.Update)
	r.DELETE("/api/actions/:id", actionHandler.Delete)

	return r, actionRepo, noteRepo
}

func TestActionHandler(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		r, _, noteRepo := setupActionTestRouter(t)

		// Create a test note first
		note := &models.CreateNoteRequest{
			Content: "Test note for action",
			Date:    time.Now(),
		}
		createdNote, err := noteRepo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		action := models.CreateActionRequest{
			NoteID:      createdNote.ID,
			Description: "Test action",
		}
		body, _ := json.Marshal(action)

		req := httptest.NewRequest(http.MethodPost, "/api/actions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Action
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.NoteID != action.NoteID {
			t.Errorf("Expected note ID %d, got %d", action.NoteID, response.NoteID)
		}
		if response.Description != action.Description {
			t.Errorf("Expected description %q, got %q", action.Description, response.Description)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		r, actionRepo, noteRepo := setupActionTestRouter(t)

		// Create a test note and action
		note := &models.CreateNoteRequest{
			Content: "Test note for action",
			Date:    time.Now(),
		}
		createdNote, err := noteRepo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		action := &models.CreateActionRequest{
			NoteID:      createdNote.ID,
			Description: "Test action for GetByID",
		}
		createdAction, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/actions/"+string(createdAction.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Action
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ID != createdAction.ID {
			t.Errorf("Expected ID %d, got %d", createdAction.ID, response.ID)
		}
	})

	t.Run("GetByNoteID", func(t *testing.T) {
		r, actionRepo, noteRepo := setupActionTestRouter(t)

		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for actions",
			Date:    time.Now(),
		}
		createdNote, err := noteRepo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		// Create test actions
		actions := []*models.CreateActionRequest{
			{NoteID: createdNote.ID, Description: "Action 1"},
			{NoteID: createdNote.ID, Description: "Action 2"},
		}

		for _, action := range actions {
			_, err := actionRepo.Create(action)
			if err != nil {
				t.Fatalf("Failed to create test action: %v", err)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/api/actions/note/"+string(createdNote.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response []models.Action
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 actions, got %d", len(response))
		}
	})

	t.Run("Update", func(t *testing.T) {
		r, actionRepo, noteRepo := setupActionTestRouter(t)

		// Create a test note and action
		note := &models.CreateNoteRequest{
			Content: "Test note for action",
			Date:    time.Now(),
		}
		createdNote, err := noteRepo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		action := &models.CreateActionRequest{
			NoteID:      createdNote.ID,
			Description: "Test action for Update",
		}
		createdAction, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}

		update := models.UpdateActionRequest{
			Completed: true,
		}
		body, _ := json.Marshal(update)

		req := httptest.NewRequest(http.MethodPut, "/api/actions/"+string(createdAction.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response models.Action
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !response.Completed {
			t.Error("Expected action to be completed")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		r, actionRepo, noteRepo := setupActionTestRouter(t)

		// Create a test note and action
		note := &models.CreateNoteRequest{
			Content: "Test note for action",
			Date:    time.Now(),
		}
		createdNote, err := noteRepo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		action := &models.CreateActionRequest{
			NoteID:      createdNote.ID,
			Description: "Test action for Delete",
		}
		createdAction, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/actions/"+string(createdAction.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
		}

		// Verify action is deleted
		_, err = actionRepo.GetByID(createdAction.ID)
		if err == nil {
			t.Error("Expected error when getting deleted action")
		}
	})
}
