package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"reward-points-ledger/internal/domain"
	"reward-points-ledger/internal/service"
	"strconv"
)

type HTTPHandler struct {
	svc *service.LedgerService
}

func NewHTTPHandler(svc *service.LedgerService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

// ErrorResponse structural blueprint
type ErrorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
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
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" || input.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid input. Name and email are required.")
		return
	}

	m, err := h.svc.CreateMember(input.Name, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	respondWithJSON(w, http.StatusCreated, m)
}

func (h *HTTPHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "memberId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	m, err := h.svc.GetMember(id)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	respondWithJSON(w, http.StatusOK, m)
}

func (h *HTTPHandler) CreateReward(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MemberID    int    `json:"member_id"`
		PointTypeID int    `json:"point_type_id"`
		Points      int    `json:"points"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed JSON request body")
		return
	}

	rw, err := h.svc.ProcessReward(input.MemberID, input.PointTypeID, input.Points, input.Description)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			respondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrInvalidPointType), errors.Is(err, domain.ErrPointsNotPositive):
			respondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrInsufficientBalance):
			respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	respondWithJSON(w, http.StatusCreated, rw)
}

func (h *HTTPHandler) GetMemberRewards(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "memberId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	entries, err := h.svc.GetRewards(id)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// If entries is nil, return empty JSON array [] instead of null
	if entries == nil {
		entries = []domain.RewardEntry{}
	}
	respondWithJSON(w, http.StatusOK, entries)
}
