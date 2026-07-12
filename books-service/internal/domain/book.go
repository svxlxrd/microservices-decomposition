package domain

import (
	"database/sql"
	"time"
)

// ========== entities ==========

// Book основная доменная модель книги для хранения в БД
type Book struct {
	ID     string `json:"id" db:"id"`
	Title  string `json:"title" db:"title"`
	Author string `json:"author" db:"author"`
	UserID string `json:"user_id" db:"user_id"`

	Description   sql.NullString  `json:"description" db:"description"`
	ISBN          sql.NullString  `json:"isbn" db:"isbn"`
	PublishedYear sql.NullInt32   `json:"published_year" db:"published_year"`
	AverageRating sql.NullFloat64 `json:"average_rating" db:"average_rating"`

	ReviewsCount int `json:"reviews_count" db:"reviews_count"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
