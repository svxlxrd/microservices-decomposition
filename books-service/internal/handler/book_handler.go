package handler

import (
	"errors"
	"net/http"
	"strconv"

	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type BookHandler struct {
	svc *service.BookService
}

func NewBookHandler(svc *service.BookService) *BookHandler {
	return &BookHandler{
		svc: svc,
	}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user is not authenticated")
		return
	}

	var req domain.CreateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	var details []domain.ErrorDetail

	if req.Title == "" {
		details = append(details, domain.ErrorDetail{
			Field:   "title",
			Message: "title is required",
		})
	}

	if req.Author == "" {
		details = append(details, domain.ErrorDetail{
			Field:   "author",
			Message: "author is required",
		})
	}

	if len(details) > 0 {
		writeValidationError(w, r, details)
		return
	}

	book, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookTitleEmpty),
			errors.Is(err, domain.ErrBookAuthorEmpty):
			writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	bookResponse, err := h.svc.GetByID(r.Context(), bookID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")
		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, bookResponse)
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 10

	var err error

	if p := r.URL.Query().Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid page")
			return
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		limit, err = strconv.Atoi(l)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid limit")
			return
		}
	}

	filter := domain.ListParams{
		Search: r.URL.Query().Get("search"),
		Sort:   r.URL.Query().Get("sort"),
		Order:  r.URL.Query().Get("order"),
		Page:   page,
		Limit:  limit,
	}

	resp, _, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	userID, ok := getUserID(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user is not authenticated")
		return
	}

	var req domain.UpdateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	book, err := h.svc.Update(r.Context(), userID, bookID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")

		case errors.Is(err, domain.ErrNotBookOwner):
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "you are not the owner of this book")

		case errors.Is(err, domain.ErrBookTitleEmpty),
			errors.Is(err, domain.ErrBookAuthorEmpty):
			writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	userID, ok := getUserID(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "user is not authenticated")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, bookID); err != nil {
		switch {
		case errors.Is(err, domain.ErrBookNotFound):
			writeError(w, r, http.StatusNotFound, "BOOK_NOT_FOUND", "book not found")

		case errors.Is(err, domain.ErrNotBookOwner):
			writeError(w, r, http.StatusForbidden, "FORBIDDEN", "you are not the owner of this book")

		default:
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
