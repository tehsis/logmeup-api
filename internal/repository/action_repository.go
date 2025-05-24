package repository

import (
	"database/sql"
	"log"
	"time"

	"github.com/tehsis/logmeup-api/internal/models"
)

type ActionRepository struct {
	db *sql.DB
}

func NewActionRepository(db *sql.DB) *ActionRepository {
	log.Printf("[ActionRepository] Initializing action repository")
	return &ActionRepository{db: db}
}

// Helper function to log database operations
func logDBOperation(operation string, details ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[ActionRepository-%s] %s | Details: %v", operation, timestamp, details)
}

// Helper function to log database errors
func logDBError(operation string, err error, details ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[ActionRepository-%s-ERROR] %s | Error: %v | Details: %v", operation, timestamp, err, details)
}

// Helper function to log successful database operations
func logDBSuccess(operation string, details ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[ActionRepository-%s-SUCCESS] %s | Details: %v", operation, timestamp, details)
}

func (r *ActionRepository) Create(action *models.CreateActionRequest) (*models.Action, error) {
	logDBOperation("Create", "Starting action creation", map[string]interface{}{
		"note_id":     action.NoteID,
		"description": action.Description,
	})

	query := `
		INSERT INTO actions (note_id, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, note_id, description, completed, created_at, updated_at
	`

	now := time.Now()
	var createdAction models.Action

	logDBOperation("Create", "Executing SQL query", query)

	err := r.db.QueryRow(
		query,
		action.NoteID,
		action.Description,
		false,
		now,
		now,
	).Scan(
		&createdAction.ID,
		&createdAction.NoteID,
		&createdAction.Description,
		&createdAction.Completed,
		&createdAction.CreatedAt,
		&createdAction.UpdatedAt,
	)

	if err != nil {
		logDBError("Create", err, "Failed to create action", map[string]interface{}{
			"note_id":     action.NoteID,
			"description": action.Description,
		})
		return nil, err
	}

	logDBSuccess("Create", "Action created successfully", map[string]interface{}{
		"action_id":   createdAction.ID,
		"note_id":     createdAction.NoteID,
		"description": createdAction.Description,
	})

	return &createdAction, nil
}

func (r *ActionRepository) GetByID(id int64) (*models.Action, error) {
	logDBOperation("GetByID", "Fetching action by ID", id)

	query := `
		SELECT id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE id = $1
	`

	var action models.Action

	logDBOperation("GetByID", "Executing SQL query", query, "ID:", id)

	err := r.db.QueryRow(query, id).Scan(
		&action.ID,
		&action.NoteID,
		&action.Description,
		&action.Completed,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logDBError("GetByID", err, "Action not found", id)
		} else {
			logDBError("GetByID", err, "Database error while fetching action", id)
		}
		return nil, err
	}

	logDBSuccess("GetByID", "Action retrieved successfully", map[string]interface{}{
		"action_id":   action.ID,
		"description": action.Description,
		"completed":   action.Completed,
	})

	return &action, nil
}

func (r *ActionRepository) GetAll() ([]*models.Action, error) {
	logDBOperation("GetAll", "Fetching all actions")

	query := `
		SELECT id, note_id, description, completed, created_at, updated_at
		FROM actions
		ORDER BY created_at DESC
	`

	logDBOperation("GetAll", "Executing SQL query", query)

	rows, err := r.db.Query(query)
	if err != nil {
		logDBError("GetAll", err, "Failed to execute query")
		return nil, err
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		var action models.Action
		err := rows.Scan(
			&action.ID,
			&action.NoteID,
			&action.Description,
			&action.Completed,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			logDBError("GetAll", err, "Failed to scan action row")
			return nil, err
		}
		actions = append(actions, &action)
	}

	if err = rows.Err(); err != nil {
		logDBError("GetAll", err, "Error occurred during row iteration")
		return nil, err
	}

	logDBSuccess("GetAll", "Actions retrieved successfully", map[string]interface{}{
		"count": len(actions),
	})

	return actions, nil
}

func (r *ActionRepository) GetByNoteID(noteID int64) ([]*models.Action, error) {
	logDBOperation("GetByNoteID", "Fetching actions by note ID", noteID)

	query := `
		SELECT id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE note_id = $1
		ORDER BY created_at DESC
	`

	logDBOperation("GetByNoteID", "Executing SQL query", query, "Note ID:", noteID)

	rows, err := r.db.Query(query, noteID)
	if err != nil {
		logDBError("GetByNoteID", err, "Failed to execute query", noteID)
		return nil, err
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		var action models.Action
		err := rows.Scan(
			&action.ID,
			&action.NoteID,
			&action.Description,
			&action.Completed,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			logDBError("GetByNoteID", err, "Failed to scan action row", noteID)
			return nil, err
		}
		actions = append(actions, &action)
	}

	if err = rows.Err(); err != nil {
		logDBError("GetByNoteID", err, "Error occurred during row iteration", noteID)
		return nil, err
	}

	logDBSuccess("GetByNoteID", "Actions retrieved successfully", map[string]interface{}{
		"note_id": noteID,
		"count":   len(actions),
	})

	return actions, nil
}

func (r *ActionRepository) Update(id int64, action *models.UpdateActionRequest) (*models.Action, error) {
	logDBOperation("Update", "Updating action", map[string]interface{}{
		"action_id": id,
		"completed": action.Completed,
	})

	query := `
		UPDATE actions
		SET completed = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, note_id, description, completed, created_at, updated_at
	`

	now := time.Now()
	var updatedAction models.Action

	logDBOperation("Update", "Executing SQL query", query)

	err := r.db.QueryRow(
		query,
		action.Completed,
		now,
		id,
	).Scan(
		&updatedAction.ID,
		&updatedAction.NoteID,
		&updatedAction.Description,
		&updatedAction.Completed,
		&updatedAction.CreatedAt,
		&updatedAction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logDBError("Update", err, "Action not found for update", id)
		} else {
			logDBError("Update", err, "Database error while updating action", map[string]interface{}{
				"action_id": id,
				"completed": action.Completed,
			})
		}
		return nil, err
	}

	logDBSuccess("Update", "Action updated successfully", map[string]interface{}{
		"action_id":  updatedAction.ID,
		"completed":  updatedAction.Completed,
		"updated_at": updatedAction.UpdatedAt,
	})

	return &updatedAction, nil
}

func (r *ActionRepository) Delete(id int64) error {
	logDBOperation("Delete", "Deleting action", id)

	query := `DELETE FROM actions WHERE id = $1`

	logDBOperation("Delete", "Executing SQL query", query, "ID:", id)

	result, err := r.db.Exec(query, id)
	if err != nil {
		logDBError("Delete", err, "Database error while deleting action", id)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logDBError("Delete", err, "Error checking rows affected", id)
		return err
	}

	if rowsAffected == 0 {
		logDBOperation("Delete", "No action found to delete", id)
	} else {
		logDBSuccess("Delete", "Action deleted successfully", map[string]interface{}{
			"action_id":     id,
			"rows_affected": rowsAffected,
		})
	}

	return nil
}
