package handlers

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type ProjectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler(projectService services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService}
}

func (h *ProjectHandler) HandleGetAllProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectService.GetAllProjects(r.Context())
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	projectResponse := []dtos.Project{}
	for _, project := range projects {
		projectResponse = append(projectResponse, dtos.ToProjectDTO(project))
	}
	writeResponse(w, http.StatusOK, "projects fetched successfully", projectResponse)
}

func (h *ProjectHandler) HandleGetProjectByID(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	project, err := h.projectService.GetProjectByID(r.Context(), projectID)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "project not found")
		} else if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid project ID")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	responseProject := dtos.ToProjectDTO(project)
	writeResponse(w, http.StatusOK, "project fetched successfully", responseProject)
}

func (h *ProjectHandler) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	createProjectRequest, err := decodeRequest[dtos.CreateProjectRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	projectDomain := dtos.ToProjectDomain(dtos.Project{
		Name:         createProjectRequest.Name,
		Description:  createProjectRequest.Description,
		Budget:       createProjectRequest.Budget,
		StartDate:    createProjectRequest.StartDate,
		EndDate:      createProjectRequest.EndDate,
		DepartmentID: createProjectRequest.DepartmentID,
	})

	projectID, err := h.projectService.CreateProject(r.Context(), projectDomain)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid project data")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusCreated, "project created successfully", dtos.CreateProjectResponse{ID: projectID})
}

func (h *ProjectHandler) HandleUpdateProject(w http.ResponseWriter, r *http.Request) {
	updateProjectRequest, err := decodeRequest[dtos.UpdateProjectRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	projectID := r.PathValue("id")
	projectDomain := dtos.ToProjectDomain(dtos.Project{
		ID:           projectID,
		Name:         updateProjectRequest.Name,
		Description:  updateProjectRequest.Description,
		Budget:       updateProjectRequest.Budget,
		StartDate:    updateProjectRequest.StartDate,
		EndDate:      updateProjectRequest.EndDate,
		DepartmentID: updateProjectRequest.DepartmentID,
	})

	err = h.projectService.UpdateProject(r.Context(), projectDomain)
	if err != nil {
		if apperr.IsUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
		} else if apperr.IsForbiddenError(err) {
			writeError(w, http.StatusForbidden, "forbidden")
		} else if apperr.IsInvalidError(err) {
			writeError(w, http.StatusBadRequest, "invalid project data")
		} else if apperr.IsNotFoundError(err) {
			writeError(w, http.StatusNotFound, "project not found")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeResponse(w, http.StatusOK, "project updated successfully", struct{}{})
}
