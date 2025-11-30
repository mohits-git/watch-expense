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

func Test_handlers_UserHandler_HandleGetAllUsers(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.UserService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful get all users",
			setupMock: func(m *mockservice.UserService) {
				users := []domain.User{
					{ID: "1", Name: "User 1", Email: "user1@example.com"},
					{ID: "2", Name: "User 2", Email: "user2@example.com"},
				}
				m.On("GetAllUsers", context.Background()).Return(users, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "users fetched successfully",
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetAllUsers", context.Background()).
					Return([]domain.User{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "internal server error",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetAllUsers", context.Background()).
					Return([]domain.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.UserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			w := httptest.NewRecorder()

			handler.HandleGetAllUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_UserHandler_HandleGetUserByID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*mockservice.UserService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:   "successful get user by id",
			userID: "user-123",
			setupMock: func(m *mockservice.UserService) {
				user := domain.User{ID: "user-123", Name: "John Doe", Email: "john@example.com"}
				m.On("GetUserByID", context.Background(), "user-123").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "user fetched successfully",
		},
		{
			name:   "user not found",
			userID: "invalid-id",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserByID", context.Background(), "invalid-id").
					Return(domain.User{}, apperr.NewAppError(apperr.ErrNotFound, "user not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "user not found",
		},
		{
			name:   "unauthorized",
			userID: "user-123",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserByID", context.Background(), "user-123").
					Return(domain.User{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:   "invalid user ID",
			userID: "invalid",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserByID", context.Background(), "invalid").
					Return(domain.User{}, apperr.NewAppError(apperr.ErrInvalid, "invalid user ID", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.UserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/users/"+tt.userID, nil)
			req.SetPathValue("id", tt.userID)
			w := httptest.NewRecorder()

			handler.HandleGetUserByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_UserHandler_HandleCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*mockservice.UserService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful create user",
			requestBody: dtos.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     domain.Employee,
			},
			setupMock: func(m *mockservice.UserService) {
				m.On("CreateUser", context.Background(), domain.User{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
					Role:     domain.Employee,
				}).Return("user-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "user created successfully",
		},
		{
			name: "unauthorized - non admin",
			requestBody: dtos.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     domain.Employee,
			},
			setupMock: func(m *mockservice.UserService) {
				m.On("CreateUser", context.Background(), domain.User{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
					Role:     domain.Employee,
				}).Return("", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "only admin can create users",
		},
		{
			name: "invalid user data",
			requestBody: dtos.CreateUserRequest{
				Name:     "",
				Email:    "invalid",
				Password: "pass",
				Role:     "invalidrole",
			},
			setupMock: func(m *mockservice.UserService) {
				m.On("CreateUser", context.Background(), domain.User{
					Name:     "",
					Email:    "invalid",
					Password: "pass",
					Role:     "invalidrole",
				}).Return("", apperr.NewAppError(apperr.ErrInvalid, "invalid user data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid user data",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"name": 1,
			},
			setupMock:      func(m *mockservice.UserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.UserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleCreateUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_UserHandler_HandleUpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    any
		setupMock      func(*mockservice.UserService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:   "successful update user",
			userID: "user-123",
			requestBody: dtos.UpdateUserRequest{
				Name:  "John Doe Updated",
				Email: "john.updated@example.com",
				Role:  domain.Manager,
			},
			setupMock: func(m *mockservice.UserService) {
				m.On("UpdateUser", context.Background(), domain.User{
					ID:    "user-123",
					Name:  "John Doe Updated",
					Email: "john.updated@example.com",
					Role:  domain.Manager,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "user updated successfully",
		},
		{
			name:   "unauthorized",
			userID: "user-123",
			requestBody: dtos.UpdateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  domain.Employee,
			},
			setupMock: func(m *mockservice.UserService) {
				m.On("UpdateUser", context.Background(), domain.User{
					ID:    "user-123",
					Name:  "John Doe",
					Email: "john@example.com",
					Role:  domain.Employee,
				}).Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "only admin can update users",
		},
		{
			name:   "invalid user data",
			userID: "user-123",
			requestBody: dtos.UpdateUserRequest{
				Name:  "",
				Email: "invalid",
				Role:  "invalidrole",
			},
			setupMock: func(m *mockservice.UserService) {
				m.On("UpdateUser", context.Background(), domain.User{
					ID:    "user-123",
					Name:  "",
					Email: "invalid",
					Role:  "invalidrole",
				}).Return(apperr.NewAppError(apperr.ErrInvalid, "invalid user data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid user data",
		},
		{
			name:   "invalid request body",
			userID: "user-123",
			requestBody: map[string]any{
				"name": 128,
			},
			setupMock:      func(m *mockservice.UserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.UserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/users/"+tt.userID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.userID)
			w := httptest.NewRecorder()

			handler.HandleUpdateUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_UserHandler_HandleDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*mockservice.UserService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:   "successful delete user",
			userID: "user-123",
			setupMock: func(m *mockservice.UserService) {
				m.On("DeleteUser", context.Background(), "user-123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "user deleted successfully",
		},
		{
			name:   "unauthorized",
			userID: "user-123",
			setupMock: func(m *mockservice.UserService) {
				m.On("DeleteUser", context.Background(), "user-123").
					Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "only admin can delete users",
		},
		{
			name:   "invalid user ID",
			userID: "invalid",
			setupMock: func(m *mockservice.UserService) {
				m.On("DeleteUser", context.Background(), "invalid").
					Return(apperr.NewAppError(apperr.ErrInvalid, "invalid user ID", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.UserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/users/"+tt.userID, nil)
			req.SetPathValue("id", tt.userID)
			w := httptest.NewRecorder()

			handler.HandleDeleteUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_UserHandler_HandleGetUserBudget(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.UserService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful get user budget",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserBudget", context.Background()).Return(5000.00, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "user budget fetched successfully",
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserBudget", context.Background()).
					Return(0.0, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "user not found",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserBudget", context.Background()).
					Return(0.0, apperr.NewAppError(apperr.ErrNotFound, "user not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "user not found",
		},
		{
			name: "forbidden",
			setupMock: func(m *mockservice.UserService) {
				m.On("GetUserBudget", context.Background()).
					Return(0.0, apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "you can only view your own budget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.UserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/users/budget", nil)
			w := httptest.NewRecorder()

			handler.HandleGetUserBudget(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}
