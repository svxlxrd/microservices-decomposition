package handler

import (
	"bookshelf/auth-service/internal/domain"
	"bookshelf/auth-service/internal/service"
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
)

type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{
		svc: svc,
	}
}

// helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID, _ := r.Context().Value(requestIDKey).(string)

	response := map[string]domain.ErrorResponse{
		"error": {
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
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

	writeJSON(w, http.StatusBadRequest, map[string]domain.ErrorResponse{
		"error": response,
	})
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func getUserID(ctx context.Context) string {
	userID := ctx.Value(userIDKey).(string)
	return userID
}
