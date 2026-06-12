package handler

import (
	"bytes"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// CORSMiddleware handles Cross-Origin Resource Sharing restrictions.
// Initializes configuration constants once on booting to optimize runtime performance.
func CORSMiddleware() func(http.Handler) http.Handler {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:8081" // Local Swagger UI port default fallback
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle browser preflight checks instantly
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
				w.WriteHeader(http.StatusOK)
				return
			}

			err := logAndRestoreBody(r)
			if err != nil {
				respondWithError(w, r, http.StatusBadRequest, err.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func logAndRestoreBody(r *http.Request) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close() // Close the original reader

	logFields := []interface{}{
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"body", string(bodyBytes),
	}

	slog.Debug("HTTP Request", logFields...)

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return nil
}
