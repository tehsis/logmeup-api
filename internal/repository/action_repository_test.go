package repository

import (
	"testing"
	"time"

	"github.com/tehsis/logmeup-api/internal/models"
	"github.com/tehsis/logmeup-api/internal/testutil"
)

func TestActionRepository(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	testutil.SetupTestSchema(t, db)

	noteRepo := NewNoteRepository(db)
	actionRepo := NewActionRepository(db)

	// Helper function to create a test note
	createTestNote := func(t *testing.T) *models.Note {
		note := &models.CreateNoteRequest{
			Content: "Test note for actions",
			Date:    time.Now(),
		}
		created, err := noteRepo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}
		return created
	}

	t.Run("Create", func(t *testing.T) {
		note := createTestNote(t)
		action := &models.CreateActionRequest{
			NoteID:      note.ID,
			Description: "Test action",
		}

		created, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create action: %v", err)
		}

		if created.ID == 0 {
			t.Error("Expected action ID to be set")
		}
		if created.NoteID != action.NoteID {
			t.Errorf("Expected note ID %d, got %d", action.NoteID, created.NoteID)
		}
		if created.Description != action.Description {
			t.Errorf("Expected description %q, got %q", action.Description, created.Description)
		}
		if created.Completed {
			t.Error("Expected action to be not completed by default")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		note := createTestNote(t)
		action := &models.CreateActionRequest{
			NoteID:      note.ID,
			Description: "Test action for GetByID",
		}
		created, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}

		retrieved, err := actionRepo.GetByID(created.ID)
		if err != nil {
			t.Fatalf("Failed to get action: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
		}
		if retrieved.Description != created.Description {
			t.Errorf("Expected description %q, got %q", created.Description, retrieved.Description)
		}
	})

	t.Run("GetByNoteID", func(t *testing.T) {
		note := createTestNote(t)
		actions := []*models.CreateActionRequest{
			{NoteID: note.ID, Description: "Action 1"},
			{NoteID: note.ID, Description: "Action 2"},
			{NoteID: note.ID, Description: "Action 3"},
		}

		// Create test actions
		for _, action := range actions {
			_, err := actionRepo.Create(action)
			if err != nil {
				t.Fatalf("Failed to create test action: %v", err)
			}
		}

		retrieved, err := actionRepo.GetByNoteID(note.ID)
		if err != nil {
			t.Fatalf("Failed to get actions by note ID: %v", err)
		}

		if len(retrieved) != 3 {
			t.Errorf("Expected 3 actions, got %d", len(retrieved))
		}
	})

	t.Run("Update", func(t *testing.T) {
		note := createTestNote(t)
		action := &models.CreateActionRequest{
			NoteID:      note.ID,
			Description: "Test action for Update",
		}
		created, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}

		update := &models.UpdateActionRequest{
			Completed: true,
		}
		updated, err := actionRepo.Update(created.ID, update)
		if err != nil {
			t.Fatalf("Failed to update action: %v", err)
		}

		if !updated.Completed {
			t.Error("Expected action to be completed")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		note := createTestNote(t)
		action := &models.CreateActionRequest{
			NoteID:      note.ID,
			Description: "Test action for Delete",
		}
		created, err := actionRepo.Create(action)
		if err != nil {
			t.Fatalf("Failed to create test action: %v", err)
		}

		err = actionRepo.Delete(created.ID)
		if err != nil {
			t.Fatalf("Failed to delete action: %v", err)
		}

		_, err = actionRepo.GetByID(created.ID)
		if err == nil {
			t.Error("Expected error when getting deleted action")
		}
	})
}
