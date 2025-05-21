package repository

import (
	"testing"
	"time"

	"github.com/tehsis/logmeup-api/internal/models"
	"github.com/tehsis/logmeup-api/internal/testutil"
)

func TestNoteRepository(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	testutil.SetupTestSchema(t, db)

	repo := NewNoteRepository(db)

	t.Run("Create", func(t *testing.T) {
		note := &models.CreateNoteRequest{
			Content: "Test note",
			Date:    time.Now(),
		}

		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		if created.ID == 0 {
			t.Error("Expected note ID to be set")
		}
		if created.Content != note.Content {
			t.Errorf("Expected content %q, got %q", note.Content, created.Content)
		}
		if !created.Date.Equal(note.Date) {
			t.Errorf("Expected date %v, got %v", note.Date, created.Date)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for GetByID",
			Date:    time.Now(),
		}
		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		// Test GetByID
		retrieved, err := repo.GetByID(created.ID)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
		}
		if retrieved.Content != created.Content {
			t.Errorf("Expected content %q, got %q", created.Content, retrieved.Content)
		}
	})

	t.Run("GetByDate", func(t *testing.T) {
		date := time.Now()
		notes := []*models.CreateNoteRequest{
			{Content: "Note 1", Date: date},
			{Content: "Note 2", Date: date},
			{Content: "Note 3", Date: date.AddDate(0, 0, 1)}, // Different date
		}

		// Create test notes
		for _, note := range notes {
			_, err := repo.Create(note)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}
		}

		// Test GetByDate
		retrieved, err := repo.GetByDate(date)
		if err != nil {
			t.Fatalf("Failed to get notes by date: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(retrieved))
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for Update",
			Date:    time.Now(),
		}
		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		// Test Update
		update := &models.UpdateNoteRequest{
			Content: "Updated content",
		}
		updated, err := repo.Update(created.ID, update)
		if err != nil {
			t.Fatalf("Failed to update note: %v", err)
		}

		if updated.Content != update.Content {
			t.Errorf("Expected content %q, got %q", update.Content, updated.Content)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a test note
		note := &models.CreateNoteRequest{
			Content: "Test note for Delete",
			Date:    time.Now(),
		}
		created, err := repo.Create(note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}

		// Test Delete
		err = repo.Delete(created.ID)
		if err != nil {
			t.Fatalf("Failed to delete note: %v", err)
		}

		// Verify note is deleted
		_, err = repo.GetByID(created.ID)
		if err == nil {
			t.Error("Expected error when getting deleted note")
		}
	})
}
