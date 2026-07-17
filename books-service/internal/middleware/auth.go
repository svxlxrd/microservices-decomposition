package middleware

import (
	"bookshelf/books-service/internal/client"
	"context"
	"errors"
	"net/http"
	"strings"
)

func AuthMiddleware(authClient *client.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "

			if !strings.HasPrefix(authHeader, prefix) {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, prefix)
			if token == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			resp, err := authClient.VerifyToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) ||
					errors.Is(err, context.Canceled) {
					http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
					return
				}

				http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
				return
			}

			if !resp.Valid {
				http.Error(w, resp.Error, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", resp.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
