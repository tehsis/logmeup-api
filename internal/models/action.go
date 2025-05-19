package models

import "time"

type Action struct {
	ID          int64     `json:"id"`
	NoteID      int64     `json:"note_id"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateActionRequest struct {
	NoteID      int64  `json:"note_id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateActionRequest struct {
	Completed bool `json:"completed"`
} 