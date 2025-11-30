package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	mockservice "github.com/mohits-git/watch-expense/tests/mock_service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_handlers_ProjectHandler_HandleGetAllProjects(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.ProjectService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful get all projects",
			setupMock: func(m *mockservice.ProjectService) {
				projects := []domain.Project{
					{ID: "1", Name: "Project A", Budget: 50000},
					{ID: "2", Name: "Project B", Budget: 75000},
				}
				m.On("GetAllProjects", context.Background()).Return(projects, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "projects fetched successfully",
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.ProjectService) {
				m.On("GetAllProjects", context.Background()).
					Return([]domain.Project{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "internal server error",
			setupMock: func(m *mockservice.ProjectService) {
				m.On("GetAllProjects", context.Background()).
					Return([]domain.Project{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ProjectService)
			tt.setupMock(mockService)

			handler := NewProjectHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
			w := httptest.NewRecorder()

			handler.HandleGetAllProjects(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ProjectHandler_HandleGetProjectByID(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		setupMock      func(*mockservice.ProjectService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful get project by id",
			projectID: "proj-123",
			setupMock: func(m *mockservice.ProjectService) {
				project := domain.Project{ID: "proj-123", Name: "Project A", Budget: 50000}
				m.On("GetProjectByID", context.Background(), "proj-123").Return(project, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "project fetched successfully",
		},
		{
			name:      "project not found",
			projectID: "invalid-id",
			setupMock: func(m *mockservice.ProjectService) {
				m.On("GetProjectByID", context.Background(), "invalid-id").
					Return(domain.Project{}, apperr.NewAppError(apperr.ErrNotFound, "project not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "project not found",
		},
		{
			name:      "unauthorized",
			projectID: "proj-123",
			setupMock: func(m *mockservice.ProjectService) {
				m.On("GetProjectByID", context.Background(), "proj-123").
					Return(domain.Project{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "invalid project ID",
			projectID: "invalid",
			setupMock: func(m *mockservice.ProjectService) {
				m.On("GetProjectByID", context.Background(), "invalid").
					Return(domain.Project{}, apperr.NewAppError(apperr.ErrInvalid, "invalid project ID", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid project ID",
		},
		{
			name:      "internal server error",
			projectID: "proj-123",
			setupMock: func(m *mockservice.ProjectService) {
				m.On("GetProjectByID", context.Background(), "proj-123").
					Return(domain.Project{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ProjectService)
			tt.setupMock(mockService)

			handler := NewProjectHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/projects/"+tt.projectID, nil)
			req.SetPathValue("id", tt.projectID)
			w := httptest.NewRecorder()

			handler.HandleGetProjectByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ProjectHandler_HandleCreateProject(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*mockservice.ProjectService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful create project",
			requestBody: dtos.CreateProjectRequest{
				Name:        "New Project",
				Description: "Project description",
				Budget:      100000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("CreateProject", context.Background(), domain.Project{
					Name:        "New Project",
					Description: "Project description",
					Budget:      100000,
				}).Return("proj-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "project created successfully",
		},
		{
			name: "unauthorized",
			requestBody: dtos.CreateProjectRequest{
				Name:   "New Project",
				Budget: 100000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("CreateProject", context.Background(), domain.Project{
					Name:   "New Project",
					Budget: 100000,
				}).Return("", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "forbidden",
			requestBody: dtos.CreateProjectRequest{
				Name:   "New Project",
				Budget: 100000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("CreateProject", context.Background(), domain.Project{
					Name:   "New Project",
					Budget: 100000,
				}).Return("", apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name: "invalid project data",
			requestBody: dtos.CreateProjectRequest{
				Name:   "",
				Budget: -1000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("CreateProject", context.Background(), domain.Project{
					Name:   "",
					Budget: -1000,
				}).Return("", apperr.NewAppError(apperr.ErrInvalid, "invalid project data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid project data",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"name": 1,
			},
			setupMock:      func(m *mockservice.ProjectService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name: "internal server error",
			requestBody: dtos.CreateProjectRequest{
				Name:   "New Project",
				Budget: 100000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("CreateProject", context.Background(), domain.Project{
					Name:   "New Project",
					Budget: 100000,
				}).Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ProjectService)
			tt.setupMock(mockService)

			handler := NewProjectHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleCreateProject(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ProjectHandler_HandleUpdateProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		requestBody    any
		setupMock      func(*mockservice.ProjectService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful update project",
			projectID: "proj-123",
			requestBody: dtos.UpdateProjectRequest{
				Name:        "Updated Project",
				Description: "Updated description",
				Budget:      150000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("UpdateProject", context.Background(), domain.Project{
					ID:          "proj-123",
					Name:        "Updated Project",
					Description: "Updated description",
					Budget:      150000,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "project updated successfully",
		},
		{
			name:      "project not found",
			projectID: "invalid-id",
			requestBody: dtos.UpdateProjectRequest{
				Name:   "Updated Project",
				Budget: 150000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("UpdateProject", context.Background(), domain.Project{
					ID:     "invalid-id",
					Name:   "Updated Project",
					Budget: 150000,
				}).Return(apperr.NewAppError(apperr.ErrNotFound, "project not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "project not found",
		},
		{
			name:      "unauthorized",
			projectID: "proj-123",
			requestBody: dtos.UpdateProjectRequest{
				Name:   "Updated Project",
				Budget: 150000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("UpdateProject", context.Background(), domain.Project{
					ID:     "proj-123",
					Name:   "Updated Project",
					Budget: 150000,
				}).Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "forbidden",
			projectID: "proj-123",
			requestBody: dtos.UpdateProjectRequest{
				Name:   "Updated Project",
				Budget: 150000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("UpdateProject", context.Background(), domain.Project{
					ID:     "proj-123",
					Name:   "Updated Project",
					Budget: 150000,
				}).Return(apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:      "invalid project data",
			projectID: "proj-123",
			requestBody: dtos.UpdateProjectRequest{
				Name:   "",
				Budget: -1000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("UpdateProject", context.Background(), domain.Project{
					ID:     "proj-123",
					Name:   "",
					Budget: -1000,
				}).Return(apperr.NewAppError(apperr.ErrInvalid, "invalid project data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid project data",
		},
		{
			name:      "invalid request body",
			projectID: "proj-123",
			requestBody: map[string]any{
				"name": 22,
			},
			setupMock:      func(m *mockservice.ProjectService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name:      "internal server error",
			projectID: "proj-123",
			requestBody: dtos.UpdateProjectRequest{
				Name:   "Updated Project",
				Budget: 150000,
			},
			setupMock: func(m *mockservice.ProjectService) {
				m.On("UpdateProject", context.Background(), domain.Project{
					ID:     "proj-123",
					Name:   "Updated Project",
					Budget: 150000,
				}).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ProjectService)
			tt.setupMock(mockService)

			handler := NewProjectHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/projects/"+tt.projectID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.projectID)
			w := httptest.NewRecorder()

			handler.HandleUpdateProject(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}
