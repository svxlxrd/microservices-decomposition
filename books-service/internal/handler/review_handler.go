package handler

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/service"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		svc: svc,
	}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "book_id")

	var req domain.CreateReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	var details []domain.ErrorDetail

	if req.Rating < 1 || req.Rating > 5 {
		details = append(details, domain.ErrorDetail{
			Field:   "rating",
			Message: "rating must be between 1 and 5",
		})
	}

	if len(req.Content) < 10 {
		details = append(details, domain.ErrorDetail{
			Field:   "content",
			Message: "content must be at least 10 characters",
		})
	}

	if len(details) > 0 {
		writeValidationError(w, r, details)
		return
	}

	review, err := h.svc.Create(r.Context(), req.UserID, bookID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")

		case errors.Is(err, domain.ErrAlreadyReviewed):
			writeError(w, r, http.StatusConflict, "ALREADY_REVIEWED", "you have already reviewed this book")

		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, review)
}

func (h *ReviewHandler) ListBookReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "book_id")

	reviewList, err := h.svc.ListByBook(r.Context(), bookID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, reviewList)
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "id")

	var req domain.UpdateReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	review, err := h.svc.Update(r.Context(), req.UserID, reviewID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReviewNotFound):
			writeError(w, r, http.StatusNotFound, "REVIEW_NOT_FOUND", "review not found")

		case errors.Is(err, domain.ErrNotReviewOwner):
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "you are not the owner of this review")

		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, review)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "id")
	var userID string

	if err := decodeJSON(r, &userID); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, reviewID); err != nil {
		switch {
		case errors.Is(err, domain.ErrReviewNotFound):
			writeError(w, r, http.StatusNotFound, "REVIEW_NOT_FOUND", "review not found")

		case errors.Is(err, domain.ErrNotReviewOwner):
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "you are not the owner of this review")

		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
