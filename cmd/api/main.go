package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"reward-points-ledger/internal/handler"
	"reward-points-ledger/internal/repository"
	"reward-points-ledger/internal/service"
	"time"
)

func main() {
	InitLogger()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://root:password@localhost:5432/rewards_db?sslmode=disable"
	}

	// Call the DB initialization function
	pool, err := repository.InitDBPool(dbURL)
	if err != nil {
		slog.Error("Fatal DB initialization crash", "error", err)
	}
	defer pool.Close()

	// 1. Dependency Initialization
	//repo := repository.NewMemoryRepository()
	repo := repository.NewPostgresRepository(pool)
	svc := service.NewLedgerService(repo)
	h := handler.NewHTTPHandler(svc)

	// 2. Router Configurations & Global Middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(handler.CORSMiddleware())

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// 3. Endpoint Mapping
	r.Post("/members", h.CreateMember)
	r.Get("/members/{memberId}", h.GetMember)
	r.Post("/rewards", h.CreateReward)
	r.Get("/members/{memberId}/rewards", h.GetMemberRewards)

	r.Get("/members", h.GetAllMembers)
	r.Get("/rewards", h.GetAllRewards)

	r.Get("/member/{memberId}/point-category", h.GetMemberWithPointCategory)

	r.Post("/points", h.CreatePoint)
	r.Get("/points/{pointTypeId}", h.GetPointDetailsByPointType)
	r.Get("/points", h.GetAllPoints)
	r.Post("/points/activate", h.ActivatePoint)

	slog.Info("Reward points ledger service running on port :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		slog.Error("Server lifecycle crash", "error", err)
	}
}
