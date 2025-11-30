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
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockservice "github.com/mohits-git/watch-expense/tests/mock_service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_handlers_AuthHandler_HandleLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*mockservice.AuthenticationService)
		expectedStatus int
		expectedMsg    string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: dtos.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("Login", context.Background(), "user@example.com", "password123").
					Return("test-token-123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "login successful",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Status  int
					Message string
					Data    dtos.LoginResponse
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "test-token-123", response.Data.Token)
				assert.Contains(t, w.Header().Get("Set-Cookie"), "token=test-token-123")
			},
		},
		{
			name: "invalid email",
			requestBody: dtos.LoginRequest{
				Email:    "invalid@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("Login", context.Background(), "invalid@example.com", "password123").
					Return("", apperr.NewAppError(apperr.ErrNotFound, "user not found", errors.New("not found")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "invalid email or password",
		},
		{
			name: "invalid password",
			requestBody: dtos.LoginRequest{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("Login", context.Background(), "user@example.com", "wrongpassword").
					Return("", apperr.NewAppError(apperr.ErrUnauthorized, "invalid password", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "invalid email or password",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"email": 234,
			},
			setupMock:      func(m *mockservice.AuthenticationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name: "invalid input",
			requestBody: dtos.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("Login", context.Background(), "", "password123").
					Return("", apperr.NewAppError(apperr.ErrInvalid, "invalid input", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid inputs",
		},
		{
			name: "internal server error",
			requestBody: dtos.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("Login", context.Background(), "user@example.com", "password123").
					Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(mockservice.AuthenticationService)
			tt.setupMock(mockAuthService)

			handler := NewAuthHandler(mockAuthService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleLogin(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			if tt.checkResponse != nil {
				body, _ = json.Marshal(tt.requestBody)
				req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
				w = httptest.NewRecorder()
				handler.HandleLogin(w, req)
				tt.checkResponse(t, w)
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AuthHandler_HandleLogout(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		setupMock      func(*mockservice.AuthenticationService)
		expectedStatus int
		expectedMsg    string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful logout",
			setupContext: func() context.Context {
				return authctx.WithToken(context.Background(), "test-token-123")
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				ctx := authctx.WithToken(context.Background(), "test-token-123")
				m.On("Logout", ctx, "test-token-123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "logout successful",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				cookie := w.Header().Get("Set-Cookie")
				assert.Contains(t, cookie, "Max-Age=0")
			},
		},
		{
			name: "no token in context",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMock:      func(m *mockservice.AuthenticationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "logout service error",
			setupContext: func() context.Context {
				return authctx.WithToken(context.Background(), "test-token-123")
			},
			setupMock: func(m *mockservice.AuthenticationService) {
				ctx := authctx.WithToken(context.Background(), "test-token-123")
				m.On("Logout", ctx, "test-token-123").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(mockservice.AuthenticationService)
			tt.setupMock(mockAuthService)

			handler := NewAuthHandler(mockAuthService)

			req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
			req = req.WithContext(tt.setupContext())
			w := httptest.NewRecorder()

			handler.HandleLogout(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AuthHandler_HandleMe(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.AuthenticationService)
		expectedStatus int
		expectedMsg    string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful get current user",
			setupMock: func(m *mockservice.AuthenticationService) {
				user := domain.User{
					ID:    "user-123",
					Name:  "John Doe",
					Email: "john@example.com",
					Role:  domain.Employee,
				}
				m.On("GetCurrentUser", context.Background()).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "user details fetched successfully",
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response struct {
					Status  int
					Message string
					Data    dtos.User
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "user-123", response.Data.ID)
				assert.Equal(t, "John Doe", response.Data.Name)
			},
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("GetCurrentUser", context.Background()).
					Return(domain.User{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "user not found",
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("GetCurrentUser", context.Background()).
					Return(domain.User{}, apperr.NewAppError(apperr.ErrNotFound, "user not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "user not found",
		},
		{
			name: "internal server error",
			setupMock: func(m *mockservice.AuthenticationService) {
				m.On("GetCurrentUser", context.Background()).
					Return(domain.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(mockservice.AuthenticationService)
			tt.setupMock(mockAuthService)

			handler := NewAuthHandler(mockAuthService)

			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			w := httptest.NewRecorder()

			handler.HandleMe(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			if tt.checkResponse != nil {
				req = httptest.NewRequest(http.MethodGet, "/api/me", nil)
				w = httptest.NewRecorder()
				handler.HandleMe(w, req)
				tt.checkResponse(t, w)
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}
