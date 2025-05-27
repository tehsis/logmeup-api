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

func (r *ActionRepository) Create(userID string, action *models.CreateActionRequest) (*models.Action, error) {
	logDBOperation("Create", "Starting action creation", map[string]interface{}{
		"user_id":     userID,
		"note_id":     action.NoteID,
		"description": action.Description,
	})

	query := `
		INSERT INTO actions (user_id, note_id, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, note_id, description, completed, created_at, updated_at
	`

	now := time.Now()
	var createdAction models.Action

	logDBOperation("Create", "Executing SQL query", query)

	err := r.db.QueryRow(
		query,
		userID,
		action.NoteID,
		action.Description,
		false,
		now,
		now,
	).Scan(
		&createdAction.ID,
		&createdAction.UserID,
		&createdAction.NoteID,
		&createdAction.Description,
		&createdAction.Completed,
		&createdAction.CreatedAt,
		&createdAction.UpdatedAt,
	)

	if err != nil {
		logDBError("Create", err, "Failed to create action", map[string]interface{}{
			"user_id":     userID,
			"note_id":     action.NoteID,
			"description": action.Description,
		})
		return nil, err
	}

	logDBSuccess("Create", "Action created successfully", map[string]interface{}{
		"action_id":   createdAction.ID,
		"user_id":     createdAction.UserID,
		"note_id":     createdAction.NoteID,
		"description": createdAction.Description,
	})

	return &createdAction, nil
}

func (r *ActionRepository) GetByID(userID string, id int64) (*models.Action, error) {
	logDBOperation("GetByID", "Fetching action by ID", id, "for user", userID)

	query := `
		SELECT id, user_id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE id = $1 AND user_id = $2
	`

	var action models.Action

	logDBOperation("GetByID", "Executing SQL query", query, "ID:", id, "UserID:", userID)

	err := r.db.QueryRow(query, id, userID).Scan(
		&action.ID,
		&action.UserID,
		&action.NoteID,
		&action.Description,
		&action.Completed,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logDBError("GetByID", err, "Action not found", id, "for user", userID)
		} else {
			logDBError("GetByID", err, "Database error while fetching action", id, "for user", userID)
		}
		return nil, err
	}

	logDBSuccess("GetByID", "Action retrieved successfully", map[string]interface{}{
		"action_id":   action.ID,
		"user_id":     action.UserID,
		"description": action.Description,
		"completed":   action.Completed,
	})

	return &action, nil
}

func (r *ActionRepository) GetAll(userID string) ([]*models.Action, error) {
	logDBOperation("GetAll", "Fetching all actions for user", userID)

	query := `
		SELECT id, user_id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	logDBOperation("GetAll", "Executing SQL query", query, "UserID:", userID)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		logDBError("GetAll", err, "Failed to execute query for user", userID)
		return nil, err
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		var action models.Action
		err := rows.Scan(
			&action.ID,
			&action.UserID,
			&action.NoteID,
			&action.Description,
			&action.Completed,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			logDBError("GetAll", err, "Failed to scan action row for user", userID)
			return nil, err
		}
		actions = append(actions, &action)
	}

	if err = rows.Err(); err != nil {
		logDBError("GetAll", err, "Error occurred during row iteration for user", userID)
		return nil, err
	}

	logDBSuccess("GetAll", "Actions retrieved successfully", map[string]interface{}{
		"user_id": userID,
		"count":   len(actions),
	})

	return actions, nil
}

func (r *ActionRepository) GetByNoteID(userID string, noteID int64) ([]*models.Action, error) {
	logDBOperation("GetByNoteID", "Fetching actions by note ID", noteID, "for user", userID)

	query := `
		SELECT id, user_id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE note_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`

	logDBOperation("GetByNoteID", "Executing SQL query", query, "Note ID:", noteID, "UserID:", userID)

	rows, err := r.db.Query(query, noteID, userID)
	if err != nil {
		logDBError("GetByNoteID", err, "Failed to execute query", noteID, "for user", userID)
		return nil, err
	}
	defer rows.Close()

	var actions []*models.Action
	for rows.Next() {
		var action models.Action
		err := rows.Scan(
			&action.ID,
			&action.UserID,
			&action.NoteID,
			&action.Description,
			&action.Completed,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			logDBError("GetByNoteID", err, "Failed to scan action row", noteID, "for user", userID)
			return nil, err
		}
		actions = append(actions, &action)
	}

	if err = rows.Err(); err != nil {
		logDBError("GetByNoteID", err, "Error occurred during row iteration", noteID, "for user", userID)
		return nil, err
	}

	logDBSuccess("GetByNoteID", "Actions retrieved successfully", map[string]interface{}{
		"user_id": userID,
		"note_id": noteID,
		"count":   len(actions),
	})

	return actions, nil
}

func (r *ActionRepository) Update(userID string, id int64, action *models.UpdateActionRequest) (*models.Action, error) {
	logDBOperation("Update", "Starting action update", map[string]interface{}{
		"user_id":   userID,
		"action_id": id,
		"completed": action.Completed,
	})

	query := `
		UPDATE actions
		SET completed = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, note_id, description, completed, created_at, updated_at
	`

	now := time.Now()
	var updatedAction models.Action

	logDBOperation("Update", "Executing SQL query", query)

	err := r.db.QueryRow(
		query,
		action.Completed,
		now,
		id,
		userID,
	).Scan(
		&updatedAction.ID,
		&updatedAction.UserID,
		&updatedAction.NoteID,
		&updatedAction.Description,
		&updatedAction.Completed,
		&updatedAction.CreatedAt,
		&updatedAction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logDBError("Update", err, "Action not found for update", id, "for user", userID)
		} else {
			logDBError("Update", err, "Failed to update action", id, "for user", userID)
		}
		return nil, err
	}

	logDBSuccess("Update", "Action updated successfully", map[string]interface{}{
		"action_id": updatedAction.ID,
		"user_id":   updatedAction.UserID,
		"completed": updatedAction.Completed,
	})

	return &updatedAction, nil
}

func (r *ActionRepository) Delete(userID string, id int64) error {
	logDBOperation("Delete", "Starting action deletion", map[string]interface{}{
		"user_id":   userID,
		"action_id": id,
	})

	query := `DELETE FROM actions WHERE id = $1 AND user_id = $2`

	logDBOperation("Delete", "Executing SQL query", query)

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		logDBError("Delete", err, "Failed to delete action", id, "for user", userID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logDBError("Delete", err, "Failed to get rows affected", id, "for user", userID)
		return err
	}

	if rowsAffected == 0 {
		logDBError("Delete", nil, "No action found to delete", id, "for user", userID)
		return sql.ErrNoRows
	}

	logDBSuccess("Delete", "Action deleted successfully", map[string]interface{}{
		"action_id":     id,
		"user_id":       userID,
		"rows_affected": rowsAffected,
	})

	return nil
}
