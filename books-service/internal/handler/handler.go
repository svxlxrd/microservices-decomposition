package handler

import (
	"bookshelf/books-service/internal/domain"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
)

// helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID, _ := r.Context().Value(requestIDKey).(string)

	response := domain.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}

	writeJSON(w, status, response)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, details []domain.ErrorDetail) {
	requestID, _ := r.Context().Value(requestIDKey).(string)

	response := domain.ErrorResponse{
		Code:      "VALIDATION_ERROR",
		Message:   "validation failed",
		Details:   details,
		RequestID: requestID,
	}

	writeJSON(w, http.StatusBadRequest, response)
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// health and ready handlers
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("failed to encode response:", err)
		return
	}
}

func ReadyHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		}); err != nil {
			log.Println("failed to encode response:", err)
		}
	}
}
