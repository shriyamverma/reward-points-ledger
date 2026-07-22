package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"reward-points-ledger/internal/domain"
	"reward-points-ledger/internal/service"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTPHandler struct {
	service *service.LedgerService
}

func NewHTTPHandler(service *service.LedgerService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

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

func respondWithServiceError(w http.ResponseWriter, r *http.Request, err error) {
	if status, msg, ok := domain.HTTPStatus(err); ok {
		respondWithError(w, r, status, msg)
		return
	}
	respondWithError(w, r, http.StatusInternalServerError, "Internal server error")
}

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func (h *HTTPHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &input); err != nil || input.Name == "" || input.Email == "" {
		respondWithError(w, r, http.StatusBadRequest, "Invalid input. Name and email are required.")
		return
	}

	m, err := h.service.CreateMember(ctx, input.Name, input.Email)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, m)
}

func (h *HTTPHandler) GetMemberByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "memberId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	m, err := h.service.GetMemberByID(ctx, id)
	if err != nil {
		respondWithServiceError(w, r, err)
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
	if err := decodeJSON(r, &input); err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Malformed JSON request body")
		return
	}

	rw, err := h.service.ProcessReward(ctx, input.MemberID, input.PointTypeID, input.Points, input.Description)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, rw)
}

func (h *HTTPHandler) GetRewardsByMemberID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "memberId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	entries, err := h.service.GetRewardsByMemberID(ctx, id)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}

	if entries == nil {
		entries = []domain.RewardEntry{}
	}
	respondWithJSON(w, http.StatusOK, entries)
}

func (h *HTTPHandler) GetAllMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	members, err := h.service.GetAllMembers(ctx)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, members)
}

func (h *HTTPHandler) GetAllRewards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rewards, err := h.service.GetAllRewards(ctx)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, rewards)
}

func (h *HTTPHandler) GetMemberWithPointCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	memberId, err := strconv.Atoi(chi.URLParam(r, "memberId"))
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid member ID format")
		return
	}

	memberWithPointCategory, err := h.service.GetMemberPointSummary(ctx, memberId)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, memberWithPointCategory)
}

func (h *HTTPHandler) CreatePoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		PointTypeID int    `json:"point_type_id"`
		PointCode   string `json:"point_code"`
	}
	if err := decodeJSON(r, &input); err != nil || input.PointTypeID < 1 || input.PointCode == "" {
		respondWithError(w, r, http.StatusBadRequest, "Invalid input.")
		return
	}

	point, err := h.service.CreatePoint(ctx, input.PointTypeID, input.PointCode)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, point)
}

func (h *HTTPHandler) GetPointDetailsByPointType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "pointTypeId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid point type ID format")
		return
	}

	point, err := h.service.GetPointDetailsByPointType(ctx, id)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, point)
}

func (h *HTTPHandler) GetAllPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	points, err := h.service.GetAllPoints(ctx)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, &points)
}

func (h *HTTPHandler) setPointActive(w http.ResponseWriter, r *http.Request, active bool) {
	ctx := r.Context()

	var input struct {
		PointTypeID int `json:"point_type_id"`
	}
	if err := decodeJSON(r, &input); err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Invalid input.")
		return
	}

	point, err := h.service.SetPointActive(ctx, input.PointTypeID, active)
	if err != nil {
		respondWithServiceError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusOK, &point)
}

func (h *HTTPHandler) ActivatePoint(w http.ResponseWriter, r *http.Request) {
	h.setPointActive(w, r, true)
}

func (h *HTTPHandler) DeactivatePoint(w http.ResponseWriter, r *http.Request) {
	h.setPointActive(w, r, false)
}
