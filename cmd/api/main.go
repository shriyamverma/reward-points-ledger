package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"reward-points-ledger/internal/handler"
	"reward-points-ledger/internal/repository"
	"reward-points-ledger/internal/service"
)

// CORS middleware handler
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from your Swagger UI origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// If it's a preflight OPTIONS request, stop early with a 200 OK
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. Dependency Initialization
	repo := repository.NewMemoryRepository()
	svc := service.NewLedgerService(repo)
	h := handler.NewHTTPHandler(svc)

	// 2. Router Configurations
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Mount CORS Middleware
	r.Use(corsMiddleware)

	// 3. Endpoint Mapping
	r.Post("/members", h.CreateMember)
	r.Get("/members/{memberId}", h.GetMember)
	r.Post("/rewards", h.CreateReward)
	r.Get("/members/{memberId}/rewards", h.GetMemberRewards)

	log.Println("Reward points ledger service running on port :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server lifecycle crash: %v", err)
	}
}
