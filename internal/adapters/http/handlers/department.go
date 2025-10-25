package handlers

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type DepartmentHandler struct {
	departmentService services.DepartmentService
}

func NewDepartmentHandler(departmentService services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{departmentService}
}

func (h *DepartmentHandler) HandleGetAllDepartments(w http.ResponseWriter, r *http.Request) {
	departments, err := h.departmentService.GetAllDepartments(r.Context())
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	responseDepartments := []dtos.Department{}
	for _, department := range departments {
		responseDepartments = append(responseDepartments, dtos.ToDepartmentDTO(department))
	}
	writeResponse(w, http.StatusOK, "departments fetched successfully", responseDepartments)
}

func (h *DepartmentHandler) HandleCreateDepartment(w http.ResponseWriter, r *http.Request) {
	createDepartmentRequest, err := decodeRequest[dtos.CreateDepartmentRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	department := domain.Department{
		Name:   createDepartmentRequest.Name,
		Budget: createDepartmentRequest.Budget,
	}

	id, err := h.departmentService.CreateDepartment(r.Context(), department)
	if err != nil {
		if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid department data")
		} else if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusCreated, "department created successfully", dtos.CreateProjectResponse{ID: id})
}

func (h *DepartmentHandler) HandleGetDepartmentByID(w http.ResponseWriter, r *http.Request) {
	departmentID := r.PathValue("id")
	department, err := h.departmentService.GetDepartmentByID(r.Context(), departmentID)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "department not found")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid department ID")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	responseDepartment := dtos.ToDepartmentDTO(department)
	writeResponse(w, http.StatusOK, "department fetched successfully", responseDepartment)
}

func (h *DepartmentHandler) HandleUpdateDepartment(w http.ResponseWriter, r *http.Request) {
	departmentID := r.PathValue("id")
	updateDepartmentRequest, err := decodeRequest[dtos.UpdateDepartmentRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	department := domain.Department{
		ID:     departmentID,
		Name:   updateDepartmentRequest.Name,
		Budget: updateDepartmentRequest.Budget,
	}

	err = h.departmentService.UpdateDepartment(r.Context(), department)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "department not found")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid department data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusOK, "department updated successfully", struct{}{})
}
