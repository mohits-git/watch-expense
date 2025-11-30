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

func Test_handlers_DepartmentHandler_HandleGetAllDepartments(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.DepartmentService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful get all departments",
			setupMock: func(m *mockservice.DepartmentService) {
				departments := []domain.Department{
					{ID: "1", Name: "Engineering", Budget: 100000},
					{ID: "2", Name: "Marketing", Budget: 50000},
				}
				m.On("GetAllDepartments", context.Background()).Return(departments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "departments fetched successfully",
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("GetAllDepartments", context.Background()).
					Return([]domain.Department{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "internal server error",
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("GetAllDepartments", context.Background()).
					Return([]domain.Department{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.DepartmentService)
			tt.setupMock(mockService)

			handler := NewDepartmentHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/departments", nil)
			w := httptest.NewRecorder()

			handler.HandleGetAllDepartments(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_DepartmentHandler_HandleCreateDepartment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*mockservice.DepartmentService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful create department",
			requestBody: dtos.CreateDepartmentRequest{
				Name:   "Engineering",
				Budget: 100000,
			},
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("CreateDepartment", context.Background(), domain.Department{
					Name:   "Engineering",
					Budget: 100000,
				}).Return("dept-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "department created successfully",
		},
		{
			name: "invalid department data",
			requestBody: dtos.CreateDepartmentRequest{
				Name:   "",
				Budget: -1000,
			},
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("CreateDepartment", context.Background(), domain.Department{
					Name:   "",
					Budget: -1000,
				}).Return("", apperr.NewAppError(apperr.ErrInvalid, "invalid department data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid department data",
		},
		{
			name: "unauthorized",
			requestBody: dtos.CreateDepartmentRequest{
				Name:   "Engineering",
				Budget: 100000,
			},
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("CreateDepartment", context.Background(), domain.Department{
					Name:   "Engineering",
					Budget: 100000,
				}).Return("", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
        "name": 123,
			},
			setupMock:      func(m *mockservice.DepartmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.DepartmentService)
			tt.setupMock(mockService)

			handler := NewDepartmentHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/departments", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleCreateDepartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_DepartmentHandler_HandleGetDepartmentByID(t *testing.T) {
	tests := []struct {
		name           string
		departmentID   string
		setupMock      func(*mockservice.DepartmentService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:         "successful get department by id",
			departmentID: "dept-123",
			setupMock: func(m *mockservice.DepartmentService) {
				department := domain.Department{ID: "dept-123", Name: "Engineering", Budget: 100000}
				m.On("GetDepartmentByID", context.Background(), "dept-123").Return(department, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "department fetched successfully",
		},
		{
			name:         "department not found",
			departmentID: "invalid-id",
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("GetDepartmentByID", context.Background(), "invalid-id").
					Return(domain.Department{}, apperr.NewAppError(apperr.ErrNotFound, "department not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "department not found",
		},
		{
			name:         "invalid department ID",
			departmentID: "invalid",
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("GetDepartmentByID", context.Background(), "invalid").
					Return(domain.Department{}, apperr.NewAppError(apperr.ErrInvalid, "invalid department ID", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid department ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.DepartmentService)
			tt.setupMock(mockService)

			handler := NewDepartmentHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/departments/"+tt.departmentID, nil)
			req.SetPathValue("id", tt.departmentID)
			w := httptest.NewRecorder()

			handler.HandleGetDepartmentByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_DepartmentHandler_HandleUpdateDepartment(t *testing.T) {
	tests := []struct {
		name           string
		departmentID   string
		requestBody    any
		setupMock      func(*mockservice.DepartmentService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:         "successful update department",
			departmentID: "dept-123",
			requestBody: dtos.UpdateDepartmentRequest{
				Name:   "Engineering Updated",
				Budget: 150000,
			},
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("UpdateDepartment", context.Background(), domain.Department{
					ID:     "dept-123",
					Name:   "Engineering Updated",
					Budget: 150000,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "department updated successfully",
		},
		{
			name:         "department not found",
			departmentID: "invalid-id",
			requestBody: dtos.UpdateDepartmentRequest{
				Name:   "Engineering",
				Budget: 100000,
			},
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("UpdateDepartment", context.Background(), domain.Department{
					ID:     "invalid-id",
					Name:   "Engineering",
					Budget: 100000,
				}).Return(apperr.NewAppError(apperr.ErrNotFound, "department not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "department not found",
		},
		{
			name:         "invalid department data",
			departmentID: "dept-123",
			requestBody: dtos.UpdateDepartmentRequest{
				Name:   "",
				Budget: -1000,
			},
			setupMock: func(m *mockservice.DepartmentService) {
				m.On("UpdateDepartment", context.Background(), domain.Department{
					ID:     "dept-123",
					Name:   "",
					Budget: -1000,
				}).Return(apperr.NewAppError(apperr.ErrInvalid, "invalid department data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid department data",
		},
		{
			name:         "invalid request body",
			departmentID: "dept-123",
			requestBody: map[string]any{
        "name": 234,
			},
			setupMock:      func(m *mockservice.DepartmentService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.DepartmentService)
			tt.setupMock(mockService)

			handler := NewDepartmentHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/departments/"+tt.departmentID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.departmentID)
			w := httptest.NewRecorder()

			handler.HandleUpdateDepartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}
