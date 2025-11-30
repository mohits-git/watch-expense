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

func Test_handlers_AdvanceHandler_HandleGetAdvances(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mockservice.AdvanceService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:        "successful get advances",
			queryParams: "?status=PENDING",
			setupMock: func(m *mockservice.AdvanceService) {
				advances := []domain.Advance{
					{ID: "1", Amount: 1000, Purpose: "Travel", Status: domain.Pending},
					{ID: "2", Amount: 2000, Purpose: "Equipment", Status: domain.Pending},
				}
				m.On("GetAllAdvances", context.Background(), domain.AdvancesFilterOptions{
					Status: domain.Pending,
					Page:   0,
					Limit:  10,
				}).Return(advances, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advances fetched successfully",
		},
		{
			name:        "get advances with pagination",
			queryParams: "?status=PENDING&page=1&limit=5",
			setupMock: func(m *mockservice.AdvanceService) {
				advances := []domain.Advance{
					{ID: "3", Amount: 3000, Purpose: "Training", Status: domain.Pending},
				}
				m.On("GetAllAdvances", context.Background(), domain.AdvancesFilterOptions{
					Status: domain.Pending,
					Page:   1,
					Limit:  5,
				}).Return(advances, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advances fetched successfully",
		},
		{
			name:        "get advances with user filter",
			queryParams: "?status=PENDING&user_id=user-123",
			setupMock: func(m *mockservice.AdvanceService) {
				advances := []domain.Advance{
					{ID: "1", UserID: "user-123", Amount: 1000, Purpose: "Travel", Status: domain.Pending},
				}
				m.On("GetAllAdvances", context.Background(), domain.AdvancesFilterOptions{
					UserID: "user-123",
					Status: domain.Pending,
					Page:   0,
					Limit:  10,
				}).Return(advances, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advances fetched successfully",
		},
		{
			name:        "unauthorized",
			queryParams: "?status=PENDING",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAllAdvances", context.Background(), domain.AdvancesFilterOptions{
					Status: domain.Pending,
					Page:   0,
					Limit:  10,
				}).Return([]domain.Advance{}, 0, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:        "forbidden",
			queryParams: "?status=PENDING",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAllAdvances", context.Background(), domain.AdvancesFilterOptions{
					Status: domain.Pending,
					Page:   0,
					Limit:  10,
				}).Return([]domain.Advance{}, 0, apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:        "internal server error",
			queryParams: "?status=PENDING",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAllAdvances", context.Background(), domain.AdvancesFilterOptions{
					Status: domain.Pending,
					Page:   0,
					Limit:  10,
				}).Return([]domain.Advance{}, 0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.AdvanceService)
			tt.setupMock(mockService)

			handler := NewAdvanceHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/advances"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleGetAdvances(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AdvanceHandler_HandleGetAdvanceByID(t *testing.T) {
	tests := []struct {
		name           string
		advanceID      string
		setupMock      func(*mockservice.AdvanceService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful get advance by id",
			advanceID: "adv-123",
			setupMock: func(m *mockservice.AdvanceService) {
				advance := domain.Advance{ID: "adv-123", Amount: 1000, Purpose: "Travel"}
				m.On("GetAdvanceByID", context.Background(), "adv-123").Return(advance, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advance fetched successfully",
		},
		{
			name:      "advance not found",
			advanceID: "invalid-id",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceByID", context.Background(), "invalid-id").
					Return(domain.Advance{}, apperr.NewAppError(apperr.ErrNotFound, "advance not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "advance not found",
		},
		{
			name:      "unauthorized",
			advanceID: "adv-123",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceByID", context.Background(), "adv-123").
					Return(domain.Advance{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "forbidden",
			advanceID: "adv-123",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceByID", context.Background(), "adv-123").
					Return(domain.Advance{}, apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:      "internal server error",
			advanceID: "adv-123",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceByID", context.Background(), "adv-123").
					Return(domain.Advance{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.AdvanceService)
			tt.setupMock(mockService)

			handler := NewAdvanceHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/advances?id="+tt.advanceID, nil)
			w := httptest.NewRecorder()

			handler.HandleGetAdvanceByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AdvanceHandler_HandleCreateAdvance(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*mockservice.AdvanceService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful create advance",
			requestBody: dtos.CreateAdvanceRequest{
				Amount:      1000.00,
				Purpose:     "Business Trip",
				Description: "Conference travel",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("CreateAdvance", context.Background(), domain.Advance{
					Amount:      1000.00,
					Purpose:     "Business Trip",
					Description: "Conference travel",
				}).Return("adv-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "advance created successfully",
		},
		{
			name: "unauthorized",
			requestBody: dtos.CreateAdvanceRequest{
				Amount:  1000.00,
				Purpose: "Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("CreateAdvance", context.Background(), domain.Advance{
					Amount:  1000.00,
					Purpose: "Business Trip",
				}).Return("", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "forbidden",
			requestBody: dtos.CreateAdvanceRequest{
				Amount:  1000.00,
				Purpose: "Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("CreateAdvance", context.Background(), domain.Advance{
					Amount:  1000.00,
					Purpose: "Business Trip",
				}).Return("", apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name: "invalid advance data",
			requestBody: dtos.CreateAdvanceRequest{
				Amount:  0,
				Purpose: "",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("CreateAdvance", context.Background(), domain.Advance{
					Amount:  0,
					Purpose: "",
				}).Return("", apperr.NewAppError(apperr.ErrInvalid, "invalid advance data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid advance data",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"amount": "data",
			},
			setupMock:      func(m *mockservice.AdvanceService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name: "internal server error",
			requestBody: dtos.CreateAdvanceRequest{
				Amount:  1000.00,
				Purpose: "Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("CreateAdvance", context.Background(), domain.Advance{
					Amount:  1000.00,
					Purpose: "Business Trip",
				}).Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.AdvanceService)
			tt.setupMock(mockService)

			handler := NewAdvanceHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/advances", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleCreateAdvance(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AdvanceHandler_HandleUpdateAdvance(t *testing.T) {
	tests := []struct {
		name           string
		advanceID      string
		requestBody    any
		setupMock      func(*mockservice.AdvanceService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful update advance",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceRequest{
				Amount:      1500.00,
				Purpose:     "Updated Business Trip",
				Description: "Updated conference travel",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvance", context.Background(), domain.Advance{
					ID:          "adv-123",
					Amount:      1500.00,
					Purpose:     "Updated Business Trip",
					Description: "Updated conference travel",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advance updated successfully",
		},
		{
			name:      "unauthorized",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceRequest{
				Amount:  1500.00,
				Purpose: "Updated Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvance", context.Background(), domain.Advance{
					ID:      "adv-123",
					Amount:  1500.00,
					Purpose: "Updated Business Trip",
				}).Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "forbidden",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceRequest{
				Amount:  1500.00,
				Purpose: "Updated Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvance", context.Background(), domain.Advance{
					ID:      "adv-123",
					Amount:  1500.00,
					Purpose: "Updated Business Trip",
				}).Return(apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:      "advance not found",
			advanceID: "invalid-id",
			requestBody: dtos.UpdateAdvanceRequest{
				Amount:  1500.00,
				Purpose: "Updated Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvance", context.Background(), domain.Advance{
					ID:      "invalid-id",
					Amount:  1500.00,
					Purpose: "Updated Business Trip",
				}).Return(apperr.NewAppError(apperr.ErrNotFound, "advance not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "advance not found",
		},
		{
			name:      "invalid advance data",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceRequest{
				Amount:  0,
				Purpose: "",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvance", context.Background(), domain.Advance{
					ID:      "adv-123",
					Amount:  0,
					Purpose: "",
				}).Return(apperr.NewAppError(apperr.ErrInvalid, "invalid advance data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid advance data",
		},
		{
			name:      "invalid request body",
			advanceID: "adv-123",
			requestBody: map[string]any{
				"amount": "data",
			},
			setupMock:      func(m *mockservice.AdvanceService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name:      "internal server error",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceRequest{
				Amount:  1500.00,
				Purpose: "Updated Business Trip",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvance", context.Background(), domain.Advance{
					ID:      "adv-123",
					Amount:  1500.00,
					Purpose: "Updated Business Trip",
				}).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.AdvanceService)
			tt.setupMock(mockService)

			handler := NewAdvanceHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/advances?id="+tt.advanceID, bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleUpdateAdvance(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AdvanceHandler_HandleUpdateAdvanceStatus(t *testing.T) {
	tests := []struct {
		name           string
		advanceID      string
		requestBody    any
		setupMock      func(*mockservice.AdvanceService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful update advance status",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvanceStatus", context.Background(), "adv-123", domain.Approved).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advance status updated successfully",
		},
		{
			name:      "unauthorized",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvanceStatus", context.Background(), "adv-123", domain.Approved).
					Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "forbidden",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvanceStatus", context.Background(), "adv-123", domain.Approved).
					Return(apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:      "advance not found",
			advanceID: "invalid-id",
			requestBody: dtos.UpdateAdvanceStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvanceStatus", context.Background(), "invalid-id", domain.Approved).
					Return(apperr.NewAppError(apperr.ErrNotFound, "advance not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "advance not found",
		},
		{
			name:      "invalid status",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceStatusRequest{
				Status: "InvalidStatus",
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvanceStatus", context.Background(), "adv-123", domain.RequestStatus("InvalidStatus")).
					Return(apperr.NewAppError(apperr.ErrInvalid, "invalid status", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid status",
		},
		{
			name:      "invalid request body",
			advanceID: "adv-123",
			requestBody: map[string]any{
				"status": 2,
			},
			setupMock:      func(m *mockservice.AdvanceService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name:      "internal server error",
			advanceID: "adv-123",
			requestBody: dtos.UpdateAdvanceStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("UpdateAdvanceStatus", context.Background(), "adv-123", domain.Approved).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.AdvanceService)
			tt.setupMock(mockService)

			handler := NewAdvanceHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/advances/"+tt.advanceID+"/status", bytes.NewBuffer(body))
			req.SetPathValue("id", tt.advanceID)
			w := httptest.NewRecorder()

			handler.HandleUpdateAdvanceStatus(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_AdvanceHandler_HandleGetAdvanceSummary(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.AdvanceService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful get advance summary",
			setupMock: func(m *mockservice.AdvanceService) {
				summary := domain.AdvanceSummary{
					Approved:   7000.00,
					Reconciled: 2000.00,
					Pending:    3000.00,
					Rejected:   500.00,
				}
				m.On("GetAdvanceSummary", context.Background()).Return(summary, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "advance summary fetched successfully",
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceSummary", context.Background()).
					Return(domain.AdvanceSummary{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "forbidden",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceSummary", context.Background()).
					Return(domain.AdvanceSummary{}, apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name: "internal server error",
			setupMock: func(m *mockservice.AdvanceService) {
				m.On("GetAdvanceSummary", context.Background()).
					Return(domain.AdvanceSummary{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.AdvanceService)
			tt.setupMock(mockService)

			handler := NewAdvanceHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/advances/summary", nil)
			w := httptest.NewRecorder()

			handler.HandleGetAdvanceSummary(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}
