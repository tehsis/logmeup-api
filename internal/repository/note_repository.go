package repository

import (
	"database/sql"
	"time"

	"github.com/tehsis/logmeup-api/internal/models"
)

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(userID string, note *models.CreateNoteRequest) (*models.Note, error) {
	query := `
		INSERT INTO notes (user_id, content, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, content, date, created_at, updated_at
	`

	now := time.Now()
	var createdNote models.Note
	err := r.db.QueryRow(
		query,
		userID,
		note.Content,
		note.Date,
		now,
		now,
	).Scan(
		&createdNote.ID,
		&createdNote.UserID,
		&createdNote.Content,
		&createdNote.Date,
		&createdNote.CreatedAt,
		&createdNote.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdNote, nil
}

func (r *NoteRepository) GetByID(userID string, id int64) (*models.Note, error) {
	query := `
		SELECT id, user_id, content, date, created_at, updated_at
		FROM notes
		WHERE id = $1 AND user_id = $2
	`

	var note models.Note
	err := r.db.QueryRow(query, id, userID).Scan(
		&note.ID,
		&note.UserID,
		&note.Content,
		&note.Date,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (r *NoteRepository) GetByDate(userID string, date time.Time) ([]*models.Note, error) {
	query := `
		SELECT id, user_id, content, date, created_at, updated_at
		FROM notes
		WHERE date = $1 AND user_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, date, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.Content,
			&note.Date,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, &note)
	}

	return notes, nil
}

func (r *NoteRepository) Update(userID string, id int64, note *models.UpdateNoteRequest) (*models.Note, error) {
	query := `
		UPDATE notes
		SET content = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, content, date, created_at, updated_at
	`

	now := time.Now()
	var updatedNote models.Note
	err := r.db.QueryRow(
		query,
		note.Content,
		now,
		id,
		userID,
	).Scan(
		&updatedNote.ID,
		&updatedNote.UserID,
		&updatedNote.Content,
		&updatedNote.Date,
		&updatedNote.CreatedAt,
		&updatedNote.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updatedNote, nil
}

func (r *NoteRepository) Delete(userID string, id int64) error {
	query := `DELETE FROM notes WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, id, userID)
	return err
}
