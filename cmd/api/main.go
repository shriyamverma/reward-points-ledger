package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"reward-points-ledger/internal/handler"
	"reward-points-ledger/internal/repository"
	"reward-points-ledger/internal/service"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://root:password@localhost:5432/rewards_db?sslmode=disable"
	}

	// Call the DB initialization function
	pool, err := repository.InitDBPool(dbURL)
	if err != nil {
		log.Fatalf("Fatal initialization crash: %v", err)
	}
	defer pool.Close()

	// 1. Dependency Initialization
	//repo := repository.NewMemoryRepository()
	repo := repository.NewPostgresRepository(pool)
	svc := service.NewLedgerService(repo)
	h := handler.NewHTTPHandler(svc)

	// 2. Router Configurations & Global Middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handler.CORSMiddleware())

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
