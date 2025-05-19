package repository

import (
	"database/sql"
	"time"

	"github.com/tehsis/logmeup-api/internal/models"
)

type ActionRepository struct {
	db *sql.DB
}

func NewActionRepository(db *sql.DB) *ActionRepository {
	return &ActionRepository{db: db}
}

func (r *ActionRepository) Create(action *models.CreateActionRequest) (*models.Action, error) {
	query := `
		INSERT INTO actions (note_id, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, note_id, description, completed, created_at, updated_at
	`

	now := time.Now()
	var createdAction models.Action
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
		return nil, err
	}

	return &createdAction, nil
}

func (r *ActionRepository) GetByID(id int64) (*models.Action, error) {
	query := `
		SELECT id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE id = $1
	`

	var action models.Action
	err := r.db.QueryRow(query, id).Scan(
		&action.ID,
		&action.NoteID,
		&action.Description,
		&action.Completed,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &action, nil
}

func (r *ActionRepository) GetByNoteID(noteID int64) ([]*models.Action, error) {
	query := `
		SELECT id, note_id, description, completed, created_at, updated_at
		FROM actions
		WHERE note_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, noteID)
	if err != nil {
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
			return nil, err
		}
		actions = append(actions, &action)
	}

	return actions, nil
}

func (r *ActionRepository) Update(id int64, action *models.UpdateActionRequest) (*models.Action, error) {
	query := `
		UPDATE actions
		SET completed = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, note_id, description, completed, created_at, updated_at
	`

	now := time.Now()
	var updatedAction models.Action
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
		return nil, err
	}

	return &updatedAction, nil
}

func (r *ActionRepository) Delete(id int64) error {
	query := `DELETE FROM actions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
} 