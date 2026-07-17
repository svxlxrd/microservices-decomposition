package middleware

import "net/http"

func ServiceKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serviceKey := r.Header.Get("X-Service-Key")
			if serviceKey == "" {
				http.Error(w, "missing service key", http.StatusUnauthorized)
				return
			}

			if serviceKey != expectedKey {
				http.Error(w, "invalid service key", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

