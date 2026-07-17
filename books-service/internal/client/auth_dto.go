package client

import "time"

// ===== VerifyToken =====
type VerifyRequest struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}


// ===== GetUsersByIDs =====
type UserPublic struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type GetUsersByIDsRequest struct {
	IDs []string `json:"ids"`
}

type GetUsersByIDsResponse struct {
	Users []UserPublic `json:"users"`
}