package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"reward-points-ledger/internal/domain"
	"reward-points-ledger/internal/service"
	"strconv"
)

type HTTPHandler struct {
	service *service.LedgerService
}

func NewHTTPHandler(service *service.LedgerService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// ErrorResponse structural blueprint
type ErrorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	logFields := []interface{}{
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"status_code", code,
		"error_message", msg,
	}
	if code >= 500 {
		slog.Error("HTTP handler request failed", logFields...)
	} else {
		slog.Warn("HTTP handler request rejected", logFields...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *HTTPHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" || input.Email == "" {
		respondWithError(w, r, http.StatusBadRequest, "Invalid input. Name and email are required.")
		return
	}

	m, err := h.service.CreateMember(ctx, input.Name, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			respondWithError(w, r, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	respondWithJSON(w, http.StatusCreated, m)
}

func (h *HTTPHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "memberId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	m, err := h.service.GetMember(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			respondWithError(w, r, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	respondWithJSON(w, http.StatusOK, m)
}

func (h *HTTPHandler) CreateReward(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		MemberID    int    `json:"member_id"`
		PointTypeID int    `json:"point_type_id"`
		Points      int    `json:"points"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Malformed JSON request body")
		return
	}

	rw, err := h.service.ProcessReward(ctx, input.MemberID, input.PointTypeID, input.Points, input.Description)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			respondWithError(w, r, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrInvalidPointType), errors.Is(err, domain.ErrPointsNotPositive):
			respondWithError(w, r, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrInsufficientBalance):
			respondWithError(w, r, http.StatusUnprocessableEntity, err.Error())
		default:
			respondWithError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	respondWithJSON(w, http.StatusCreated, rw)
}

func (h *HTTPHandler) GetMemberRewards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "memberId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	entries, err := h.service.GetRewards(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			respondWithError(w, r, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	// If entries is nil, return empty JSON array [] instead of null
	if entries == nil {
		entries = []domain.RewardEntry{}
	}
	respondWithJSON(w, http.StatusOK, entries)
}
