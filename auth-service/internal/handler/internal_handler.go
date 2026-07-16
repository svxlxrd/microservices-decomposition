package handler

import (
	"bookshelf/auth-service/internal/service"
	"net/http"
	"time"
)

type InternalHandler struct {
	svc *service.UserService
}

func NewInternalHandler(svc *service.UserService) *InternalHandler {
	return &InternalHandler{
		svc: svc,
	}
}

// DTO
type VerifyRequest struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}

func (h *InternalHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	claims, err := h.svc.ValidateToken(req.Token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, VerifyResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, VerifyResponse{
		Valid:     true,
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt,
	})
}
