package models

import "time"

type Note struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateNoteRequest struct {
	Content string    `json:"content" binding:"required"`
	Date    time.Time `json:"date" binding:"required"`
}

type UpdateNoteRequest struct {
	Content string `json:"content" binding:"required"`
} 