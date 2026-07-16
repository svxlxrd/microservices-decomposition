package domain

import "errors"

var (
	// Review errors
	ErrReviewNotFound        = errors.New("review not found")
	ErrNotReviewOwner        = errors.New("not review owner")
	ErrAlreadyReviewed       = errors.New("already reviewed")
	ErrInvalidRating         = errors.New("invalid rating")
	ErrReviewContentTooShort = errors.New("review content too short")

	// Book errors
	ErrBookNotFound    = errors.New("book not found")
	ErrNotBookOwner    = errors.New("not book owner")
	ErrBookTitleEmpty  = errors.New("book title empty")
	ErrBookAuthorEmpty = errors.New("book author empty")
)