package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"bookshelf/books-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewrepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{
		db: db,
	}
}

func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	review.ID = uuid.New().String()

	query := `
	INSERT INTO reviews (id, book_id, user_id, rating, title, content)
	VALUES (:id, :book_id, :user_id, :rating, :title, :content)
	RETURNING created_at, updated_at;`

	rows, err := r.db.NamedQueryContext(ctx, query, review)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&review.CreatedAt, &review.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan timestamps: %w", err)
		}
	}

	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	var review domain.Review

	query := `
	SELECT
		id,
		book_id,
		user_id,
		rating,
		title,
		content,
		created_at,
		updated_at
	FROM reviews
	WHERE id = $1;`

	err := r.db.GetContext(ctx, &review, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Review not found")
		}
		return nil, fmt.Errorf("failed to get review by ID: %w", err)
	}

	return &review, nil
}

func (r *ReviewRepository) ListByBookID(ctx context.Context, bookID string, page, limit int) ([]domain.Review, int, error) {
	var totalCount int

	totalCountQuery := `
	SELECT COUNT(*)
	FROM reviews
	WHERE book_id = $1;`

	if err := r.db.GetContext(ctx, &totalCount, totalCountQuery, bookID); err != nil {
		return nil, 0, fmt.Errorf("failed to count reviews %w", err)
	}

	// основной запрос
	query := `
	SELECT 
		id,
		book_id,
		user_id,
		rating,
		title,
		content,
		created_at,
		updated_at
	FROM reviews
	WHERE book_id = $1
	ORDER BY created_at DESC, id DESC
	LIMIT $2 OFFSET $3;`

	var reviews []domain.Review
	offset := (page - 1) * limit

	if err := r.db.SelectContext(ctx, &reviews, query, bookID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list reviews by book ID: %w", err)
	}

	return reviews, totalCount, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *domain.Review) error {
	query := `
	UPDATE reviews
	SET 
		rating = $1,
		title = $2,
		content = $3,
		updated_at = NOW()
	WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query,
		review.Rating,
		review.Title,
		review.Content,
		review.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reviews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("review with id %s not found", id)
	}

	return nil
}
