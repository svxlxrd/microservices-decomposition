package service

import (
	"context"
	"database/sql"
	"errors"

	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
)

var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrNotReviewOwner        = errors.New("not review owner")
	ErrAlreadyReviewed       = errors.New("already reviewed")
	ErrInvalidRating         = errors.New("invalid rating")
	ErrReviewContentTooShort = errors.New("review content too short")
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	bookRepo   *repository.BookRepository
}

func NewReviewService(reviewRepo *repository.ReviewRepository, bookRepo *repository.BookRepository)*ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		bookRepo: bookRepo,
	}
}

func (s *ReviewService) Create(ctx context.Context, userID string, bookID string, req domain.CreateReviewRequest) (*domain.Review, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	if book == nil {
		return nil, ErrBookNotFound
	}


	exists, err := s.reviewRepo.UserHasReviewedBook(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrAlreadyReviewed
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}

	if len(req.Content) < 10 {
		return nil, ErrReviewContentTooShort
	}

	review := &domain.Review{
		UserID:  userID,
		BookID:  bookID,
		Rating:  req.Rating,
		Content: req.Content,
	}

	if req.Title != nil {
		review.Title = sql.NullString{
			String: *req.Title,
			Valid:  true,
		}
	}

	err = s.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if review == nil {
		return nil, ErrReviewNotFound
	}

	return review, nil
}

func (s *ReviewService) ListByBook(ctx context.Context, bookID string) ([]domain.Review, error) {
	reviews, err := s.reviewRepo.ListByBookID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	return reviews, nil
}

func (s *ReviewService) Update(ctx context.Context, userID string, id string, req domain.UpdateReviewRequest) (*domain.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	if review == nil {
		return nil, ErrReviewNotFound
	}

	if review.UserID != userID {
		return nil, ErrNotReviewOwner
	}

	if req.Rating != nil {

		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, ErrInvalidRating
		}

		review.Rating = *req.Rating
	}

	if req.Content != nil {

		if len(*req.Content) < 10 {
			return nil, ErrReviewContentTooShort
		}

		review.Content = *req.Content
	}

	if req.Title != nil {

		review.Title = sql.NullString{
			String: *req.Title,
			Valid:  true,
		}
	}

	err = s.reviewRepo.Update(ctx, review)

	if err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) Delete(ctx context.Context, userID string, id string) error {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if review == nil {
		return ErrReviewNotFound
	}

	if review.UserID != userID {
		return ErrNotReviewOwner
	}

	return s.reviewRepo.Delete(ctx, id)
}


