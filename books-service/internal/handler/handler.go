package handler

import (
	contextkeys "bookshelf/books-service/internal/context"
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/dto"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// helper functions

func getUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextkeys.UserID).(string)
	return userID, ok
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID, _ := r.Context().Value(contextkeys.RequestID).(string)

	response := domain.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}

	writeJSON(w, status, response)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, details []domain.ErrorDetail) {
	requestID, _ := r.Context().Value(contextkeys.RequestID).(string)

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

// health
type HealthHandler struct {
	db      *sqlx.DB
	service string
	version string
}

func NewHealthHandler(db *sqlx.DB, service string, version string) *HealthHandler {
	return &HealthHandler{
		db:      db,
		service: service,
		version: version,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {

	checks := map[string]dto.Check{
		"database": h.checkDatabase(),
	}

	status := "ok"

	for _, check := range checks {
		if check.Status == "error" {
			status = "unhealthy"
			break
		}
	}

	response := dto.HealthResponse{
		Status:    status,
		Service:   h.service,
		Version:   h.version,
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	code := http.StatusOK

	if status == "unhealthy" {
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, response)
}

func (h *HealthHandler) checkDatabase() dto.Check {
	start := time.Now()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	err := h.db.PingContext(ctx)

	duration := time.Since(start).String()

	if err != nil {
		return dto.Check{
			Status:   "error",
			Duration: duration,
			Error:    err.Error(),
		}
	}

	return dto.Check{
		Status:   "ok",
		Duration: duration,
	}
}
