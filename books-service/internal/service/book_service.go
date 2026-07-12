package service

import (
	"context"
	"database/sql"
	"errors"

	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
)

var (
	ErrBookNotFound    = errors.New("Book not found")
	ErrNotBookOwner    = errors.New("Not book owner")
	ErrBookTitleEmpty  = errors.New("Book title empty")
	ErrBookAuthorEmpty = errors.New("Book author empty")
)

type BookService struct {
	bookRepo *repository.BookRepository
}

func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{
		bookRepo: repo,
	}
}

func (s *BookService) Create(ctx context.Context, userID string, req domain.CreateBookRequest) (*domain.Book, error) {
	if req.Author == "" {
		return nil, ErrBookAuthorEmpty
	}

	if req.Title == "" {
		return nil, ErrBookTitleEmpty
	}

	book := &domain.Book{
		Title:  req.Title,
		Author: req.Author,
		UserID: userID,
	}

	if req.ISBN != nil {
		book.ISBN = sql.NullString{
			String: *req.ISBN,
			Valid:  true,
		}
	}

	if req.PublishedYear != nil {
		book.PublishedYear = sql.NullInt32{
			Int32: int32(*req.PublishedYear),
			Valid: true,
		}
	}

	if err := s.bookRepo.Create(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookService) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if book == nil {
		return nil, ErrBookNotFound
	}

	return book, nil
}

func (s *BookService) List(ctx context.Context, params domain.ListParams) ([]domain.Book, int, error) {

	if params.Page <= 0 {
		params.Page = 1
	}

	if params.Limit <= 0 {
		params.Limit = 10
	}

	if params.Sort == "" {
		params.Sort = "created_at"
	}

	if params.Order == "" {
		params.Order = "desc"
	}

	books, total, err := s.bookRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) ListByUser(ctx context.Context, userID string, params domain.ListParams) ([]domain.Book, int, error) {

	if params.Page <= 0 {
		params.Page = 1
	}

	if params.Limit <= 0 {
		params.Limit = 10
	}

	if params.Sort == "" {
		params.Sort = "created_at"
	}

	if params.Order == "" {
		params.Order = "desc"
	}

	books, total, err := s.bookRepo.ListByUserID(ctx, userID, params)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (s *BookService) Update(ctx context.Context, userID string, bookID string, req domain.UpdateBookRequest) (*domain.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	if book == nil {
		return nil, ErrBookNotFound
	}

	if book.UserID != userID {
		return nil, ErrNotBookOwner
	}

	if req.Title != nil {
		if *req.Title == "" {
			return nil, ErrBookTitleEmpty
		}

		book.Title = *req.Title
	}

	if req.Author != nil {
		if *req.Author == "" {
			return nil, ErrBookAuthorEmpty
		}

		book.Author = *req.Author
	}

	if req.Description != nil {
		book.Description = sql.NullString{
			String: *req.Description,
			Valid:  true,
		}
	}

	if req.ISBN != nil {
		book.ISBN = sql.NullString{
			String: *req.ISBN,
			Valid:  true,
		}
	}

	if req.PublishedYear != nil {
		book.PublishedYear = sql.NullInt32{
			Int32: int32(*req.PublishedYear),
			Valid: true,
		}
	}

	err = s.bookRepo.Update(ctx, book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookService) Delete(ctx context.Context, userID string, bookID string) error {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return err
	}

	if book == nil {
		return ErrBookNotFound
	}

	if book.UserID != userID {
		return ErrNotBookOwner
	}

	err = s.bookRepo.Delete(ctx, bookID)
	if err != nil {
		return err
	}

	return nil
}
