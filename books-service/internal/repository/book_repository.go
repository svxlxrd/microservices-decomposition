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

var (
	ErrBookNotFound = errors.New("Book not found")
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) Create(ctx context.Context, book *domain.Book) error {
	book.ID = uuid.New().String()

	query := `
	INSERT INTO books (id, title, author, isbn, published_year, created_by)
	VALUES (:id, :title, :author, :isbn, :published_year, :created_by)
	RETURNING created_at, updated_at;`

	rows, err := r.db.NamedQueryContext(ctx, query, book)
	if err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan timestamps: %w", err)
		}
	}

	return nil
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	query := `
	SELECT 
		b.id, b.title, b.author, b.created_by, b.description, 
		b.isbn, b.published_year, b.created_at, b.updated_at,
		COALESCE(COUNT(r.id), 0) AS reviews_count,
		COALESCE(AVG(r.rating), 0) AS average_rating
	FROM books b
	LEFT JOIN reviews r ON b.id = r.book_id
	WHERE b.id = $1
	GROUP BY b.id`

	book := &domain.Book{}

	err := r.db.GetContext(ctx, book, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookNotFound
		}

		return nil, fmt.Errorf("get book by id: %w", err)
	}

	return book, nil
}

func (r *BookRepository) List(ctx context.Context, params domain.ListParams) ([]domain.Book, int, error) {
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	search := "%" + params.Search + "%"

	query := `
	SELECT 
		b.id, b.title, b.author, b.created_by, b.description, 
		b.isbn, b.published_year, b.created_at, b.updated_at,
		COALESCE(COUNT(r.id), 0) AS reviews_count,
		COALESCE(AVG(r.rating), 0) AS average_rating,
		COUNT(*) OVER() as total_count
	FROM books b
	LEFT JOIN reviews r ON b.id = r.book_id
	WHERE b.title ILIKE $1 OR b.author ILIKE $1
	GROUP BY b.id
	ORDER BY b.created_at DESC
	LIMIT $2 OFFSET $3
	`

	type row struct {
		domain.Book
		TotalCount int `db:"total_count"`
	}

	var rows []row

	err := r.db.SelectContext(ctx, &rows, query, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(rows) == 0 {
		return []domain.Book{}, 0, nil
	}

	books := make([]domain.Book, len(rows))
	for i, r := range rows {
		books[i] = r.Book
	}

	totalCount := rows[0].TotalCount

	return books, totalCount, nil
}

func (r *BookRepository) ListByUserID(ctx context.Context, userID string, params domain.ListParams) ([]domain.Book, int, error) {

	offset := (params.Page - 1) * params.Limit

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM books
		WHERE user_id = $1
	`

	if err := r.db.GetContext(ctx, &total, countQuery, userID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id,
			title,
			author,
			description,
			isbn,
			published_year,
			user_id,
			created_at,
			updated_at
		FROM books
		WHERE user_id = $1
	`

	args := []any{userID}

	if params.Search != "" {
		query += ` AND (title ILIKE $2 OR author ILIKE $2)`
		args = append(args, "%"+params.Search+"%")
	}

	orderBy := "created_at"

	switch params.Sort {
	case "title":
		orderBy = "title"
	case "author":
		orderBy = "author"
	case "published_year":
		orderBy = "published_year"
	}

	order := "DESC"
	if params.Order == "asc" {
		order = "ASC"
	}

	query += fmt.Sprintf(
		" ORDER BY %s %s LIMIT $%d OFFSET $%d",
		orderBy,
		order,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, params.Limit, offset)

	var books []domain.Book

	if err := r.db.SelectContext(ctx, &books, query, args...); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *BookRepository) Update(ctx context.Context, book *domain.Book) error {
	query := `
    UPDATE books 
    SET 
		title = $1, 
		author = $2, 
		description = $3, 
		isbn = $4, 
		published_year = $5, 
		updated_at = NOW()
    WHERE id = $6`

	result, err := r.db.ExecContext(ctx, query,
		book.Title,
		book.Author,
		book.Description,
		book.ISBN,
		book.PublishedYear,
		book.ID,
	)
	if err != nil {
		return fmt.Errorf("update: failed to execute query: %w", err)
	}

	// Проверяем, была ли обновлена хоть одна запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrBookNotFound
	}

	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM books WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("book with id %s not found", id)
	}

	return nil
}
