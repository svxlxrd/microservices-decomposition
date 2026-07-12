package domain

import (
	"database/sql"
	"time"
)

// ========== entities ==========

// Review основная доменная модель рецензии для хранения в БД
type Review struct {
	ID      string         `json:"id" db:"id"`
	BookID  string         `json:"book_id" db:"book_id"`
	UserID  string         `json:"user_id" db:"user_id"`
	Rating  int            `json:"rating" db:"rating"`
	Title   sql.NullString `json:"title" db:"title"`
	Content string         `json:"content" db:"content"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
