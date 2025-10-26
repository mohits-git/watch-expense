package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/validator"
)

type AdvanceHandler struct {
	advanceService services.AdvanceService
}

func NewAdvanceHandler(advanceService services.AdvanceService) *AdvanceHandler {
	return &AdvanceHandler{advanceService}
}

func (h *AdvanceHandler) HandleGetAdvances(w http.ResponseWriter, r *http.Request) {
	filterOptions, err := parseAdvancesFilterOptions(r)
	if err != nil {
		if errors.Is(err, errors.New("invalid status")) {
			writeError(w, http.StatusBadRequest, "invalid query parameters")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	advances, total, err := h.advanceService.GetAllAdvances(r.Context(), filterOptions)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response := dtos.GetAdvancesResponse{
		Advances:      dtos.ToAdvancesDTOs(advances),
		TotalAdvances: total,
	}
	writeResponse(w, http.StatusOK, "advances fetched successfully", response)
}

func (h *AdvanceHandler) HandleGetAdvanceByID(w http.ResponseWriter, r *http.Request) {
	advanceID := r.URL.Query().Get("id")
	advance, err := h.advanceService.GetAdvanceByID(r.Context(), advanceID)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "advance not found")
		} else if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	response := dtos.ToAdvanceDTO(advance)
	writeResponse(w, http.StatusOK, "advance fetched successfully", response)
}

func (h *AdvanceHandler) HandleCreateAdvance(w http.ResponseWriter, r *http.Request) {
	createAdvanceRequest, err := decodeRequest[dtos.CreateAdvanceRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	advanceID, err := h.advanceService.CreateAdvance(r.Context(), domain.Advance{
		Amount:      createAdvanceRequest.Amount,
		Purpose:     createAdvanceRequest.Purpose,
		Description: createAdvanceRequest.Description,
	})

	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid advance data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	createAdvanceResponse := dtos.CreateAdvanceResponse{ID: advanceID}
	writeResponse(w, http.StatusCreated, "advance created successfully", createAdvanceResponse)
}

func (h *AdvanceHandler) HandleUpdateAdvance(w http.ResponseWriter, r *http.Request) {
	updateAdvanceRequest, err := decodeRequest[dtos.UpdateAdvanceRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	advanceID := r.URL.Query().Get("id")
	err = h.advanceService.UpdateAdvance(r.Context(), domain.Advance{
		ID:          advanceID,
		Amount:      updateAdvanceRequest.Amount,
		Purpose:     updateAdvanceRequest.Purpose,
		Description: updateAdvanceRequest.Description,
	})
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "advance not found")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid advance data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	writeResponse(w, http.StatusOK, "advance updated successfully", struct{}{})
}

func (h *AdvanceHandler) HandleUpdateAdvanceStatus(w http.ResponseWriter, r *http.Request) {
	updateAdvanceStatusRequest, err := decodeRequest[dtos.UpdateAdvanceStatusRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	advanceID := r.PathValue("id")
	err = h.advanceService.UpdateAdvanceStatus(r.Context(), advanceID, updateAdvanceStatusRequest.Status)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "advance not found")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid status")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	writeResponse(w, http.StatusOK, "advance status updated successfully", struct{}{})
}

func parseAdvancesFilterOptions(r *http.Request) (domain.AdvancesFilterOptions, error) {
	var err error
	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	userID := r.URL.Query().Get("user_id")

	page := 0
	limit := 10

	if !validator.ValidateAdvanceStatus(domain.RequestStatus(status)) {
		err = errors.New("invalid status")
	}
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
	}
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
	}
	if err != nil {
		return domain.AdvancesFilterOptions{}, err
	}

	return domain.AdvancesFilterOptions{
		UserID: userID,
		Status: domain.RequestStatus(status),
		Page:   page,
		Limit:  limit,
	}, nil
}

func (h *AdvanceHandler) HandleGetAdvanceSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.advanceService.GetAdvanceSummary(r.Context())
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	summaryDTO := dtos.ToAdvanceSummaryDTO(summary)
	writeResponse(w, http.StatusOK, "advance summary fetched successfully", summaryDTO)
}
