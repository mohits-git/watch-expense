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

func Test_handlers_ExpenseHandler_HandleCreateExpense(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		setupMock      func(*mockservice.ExpenseService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful create expense",
			requestBody: dtos.CreateExpenseRequest{
				Amount:      500.00,
				Purpose:     "Travel",
				Description: "Business trip",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("CreateExpense", context.Background(), domain.Expense{
					Amount:      500.00,
					Purpose:     "Travel",
					Description: "Business trip",
					Bills:       []domain.Bill{},
				}).Return("exp-123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "expense created successfully",
		},
		{
			name: "unauthorized",
			requestBody: dtos.CreateExpenseRequest{
				Amount:  500.00,
				Purpose: "Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("CreateExpense", context.Background(), domain.Expense{
					Amount:  500.00,
					Purpose: "Travel",
          Bills: []domain.Bill{},
				}).Return("", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "forbidden",
			requestBody: dtos.CreateExpenseRequest{
				Amount:  500.00,
				Purpose: "Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("CreateExpense", context.Background(), domain.Expense{
					Amount:  500.00,
					Purpose: "Travel",
          Bills: []domain.Bill{},
				}).Return("", apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name: "invalid expense data",
			requestBody: dtos.CreateExpenseRequest{
				Amount:  0,
				Purpose: "",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("CreateExpense", context.Background(), domain.Expense{
					Amount:  0,
					Purpose: "",
          Bills: []domain.Bill{},
				}).Return("", apperr.NewAppError(apperr.ErrInvalid, "invalid expense data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid expense data",
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"amount": "data",
			},
			setupMock:      func(m *mockservice.ExpenseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name: "internal server error",
			requestBody: dtos.CreateExpenseRequest{
				Amount:  500.00,
				Purpose: "Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("CreateExpense", context.Background(), domain.Expense{
					Amount:  500.00,
					Purpose: "Travel",
          Bills: []domain.Bill{},
				}).Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ExpenseService)
			tt.setupMock(mockService)

			handler := NewExpenseHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/expenses", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleCreateExpense(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ExpenseHandler_HandleGetExpenses(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mockservice.ExpenseService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:        "successful get expenses",
			queryParams: "",
			setupMock: func(m *mockservice.ExpenseService) {
				expenses := []domain.Expense{
					{ID: "1", Amount: 100, Purpose: "Travel"},
					{ID: "2", Amount: 200, Purpose: "Supplies"},
				}
				m.On("GetAllExpenses", context.Background(), domain.ExpensesFilterOptions{
					Page:  0,
					Limit: 10,
				}).Return(expenses, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expenses fetched successfully",
		},
		{
			name:        "get expenses with pagination",
			queryParams: "?page=1&limit=5",
			setupMock: func(m *mockservice.ExpenseService) {
				expenses := []domain.Expense{
					{ID: "3", Amount: 300, Purpose: "Equipment"},
				}
				m.On("GetAllExpenses", context.Background(), domain.ExpensesFilterOptions{
					Page:  1,
					Limit: 5,
				}).Return(expenses, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expenses fetched successfully",
		},
		{
			name:        "get expenses with status filter",
			queryParams: "?status=PENDING",
			setupMock: func(m *mockservice.ExpenseService) {
				expenses := []domain.Expense{
					{ID: "1", Amount: 100, Purpose: "Travel", Status: domain.Pending},
				}
				m.On("GetAllExpenses", context.Background(), domain.ExpensesFilterOptions{
					Status: domain.Pending,
					Page:   0,
					Limit:  10,
				}).Return(expenses, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expenses fetched successfully",
		},
		{
			name:        "unauthorized",
			queryParams: "",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetAllExpenses", context.Background(), domain.ExpensesFilterOptions{
					Page:  0,
					Limit: 10,
				}).Return([]domain.Expense{}, 0, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:        "forbidden",
			queryParams: "",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetAllExpenses", context.Background(), domain.ExpensesFilterOptions{
					Page:  0,
					Limit: 10,
				}).Return([]domain.Expense{}, 0, apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:        "internal server error",
			queryParams: "",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetAllExpenses", context.Background(), domain.ExpensesFilterOptions{
					Page:  0,
					Limit: 10,
				}).Return([]domain.Expense{}, 0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ExpenseService)
			tt.setupMock(mockService)

			handler := NewExpenseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/expenses"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleGetExpenses(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ExpenseHandler_HandleGetExpenseByID(t *testing.T) {
	tests := []struct {
		name           string
		expenseID      string
		setupMock      func(*mockservice.ExpenseService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful get expense by id",
			expenseID: "exp-123",
			setupMock: func(m *mockservice.ExpenseService) {
				expense := domain.Expense{ID: "exp-123", Amount: 500, Purpose: "Travel"}
				m.On("GetExpenseByID", context.Background(), "exp-123").Return(expense, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expense fetched successfully",
		},
		{
			name:      "unauthorized",
			expenseID: "exp-123",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseByID", context.Background(), "exp-123").
					Return(domain.Expense{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "expense not found",
			expenseID: "invalid-id",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseByID", context.Background(), "invalid-id").
					Return(domain.Expense{}, apperr.NewAppError(apperr.ErrNotFound, "expense not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "expense not found",
		},
		{
			name:      "invalid expense ID",
			expenseID: "invalid",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseByID", context.Background(), "invalid").
					Return(domain.Expense{}, apperr.NewAppError(apperr.ErrInvalid, "invalid expense ID", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid expense ID",
		},
		{
			name:      "internal server error",
			expenseID: "exp-123",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseByID", context.Background(), "exp-123").
					Return(domain.Expense{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ExpenseService)
			tt.setupMock(mockService)

			handler := NewExpenseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/expenses/"+tt.expenseID, nil)
			req.SetPathValue("id", tt.expenseID)
			w := httptest.NewRecorder()

			handler.HandleGetExpenseByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ExpenseHandler_HandleUpdateExpense(t *testing.T) {
	tests := []struct {
		name           string
		expenseID      string
		requestBody    any
		setupMock      func(*mockservice.ExpenseService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful update expense",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseRequest{
				Amount:      600.00,
				Purpose:     "Updated Travel",
				Description: "Updated description",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpense", context.Background(), domain.Expense{
					ID:          "exp-123",
					Amount:      600.00,
					Purpose:     "Updated Travel",
					Description: "Updated description",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expense updated successfully",
		},
		{
			name:      "unauthorized",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseRequest{
				Amount:  600.00,
				Purpose: "Updated Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpense", context.Background(), domain.Expense{
					ID:      "exp-123",
					Amount:  600.00,
					Purpose: "Updated Travel",
				}).Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "forbidden",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseRequest{
				Amount:  600.00,
				Purpose: "Updated Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpense", context.Background(), domain.Expense{
					ID:      "exp-123",
					Amount:  600.00,
					Purpose: "Updated Travel",
				}).Return(apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:      "expense not found",
			expenseID: "invalid-id",
			requestBody: dtos.UpdateExpenseRequest{
				Amount:  600.00,
				Purpose: "Updated Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpense", context.Background(), domain.Expense{
					ID:      "invalid-id",
					Amount:  600.00,
					Purpose: "Updated Travel",
				}).Return(apperr.NewAppError(apperr.ErrNotFound, "expense not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "expense not found",
		},
		{
			name:      "invalid expense data",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseRequest{
				Amount:  0,
				Purpose: "",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpense", context.Background(), domain.Expense{
					ID:      "exp-123",
					Amount:  0,
					Purpose: "",
				}).Return(apperr.NewAppError(apperr.ErrInvalid, "invalid expense data", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid expense data",
		},
		{
			name:      "invalid request body",
			expenseID: "exp-123",
			requestBody: map[string]any{
				"amount": "data",
			},
			setupMock:      func(m *mockservice.ExpenseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name:      "internal server error",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseRequest{
				Amount:  600.00,
				Purpose: "Updated Travel",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpense", context.Background(), domain.Expense{
					ID:      "exp-123",
					Amount:  600.00,
					Purpose: "Updated Travel",
				}).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ExpenseService)
			tt.setupMock(mockService)

			handler := NewExpenseHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/expenses/"+tt.expenseID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.expenseID)
			w := httptest.NewRecorder()

			handler.HandleUpdateExpense(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ExpenseHandler_HandleUpdateExpenseStatus(t *testing.T) {
	tests := []struct {
		name           string
		expenseID      string
		requestBody    any
		setupMock      func(*mockservice.ExpenseService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "successful update expense status",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpenseStatus", context.Background(), "exp-123", domain.Approved).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expense status updated successfully",
		},
		{
			name:      "unauthorized",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpenseStatus", context.Background(), "exp-123", domain.Approved).
					Return(apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:      "forbidden",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpenseStatus", context.Background(), "exp-123", domain.Approved).
					Return(apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:      "expense not found",
			expenseID: "invalid-id",
			requestBody: dtos.UpdateExpenseStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpenseStatus", context.Background(), "invalid-id", domain.Approved).
					Return(apperr.NewAppError(apperr.ErrNotFound, "expense not found", errors.New("not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "expense not found",
		},
		{
			name:      "invalid status",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseStatusRequest{
				Status: "InvalidStatus",
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpenseStatus", context.Background(), "exp-123", domain.RequestStatus("InvalidStatus")).
					Return(apperr.NewAppError(apperr.ErrInvalid, "invalid status", errors.New("invalid")))
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid status",
		},
		{
			name:      "invalid request body",
			expenseID: "exp-123",
			requestBody: map[string]any{
				"status": 12,
			},
			setupMock:      func(m *mockservice.ExpenseService) {},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request",
		},
		{
			name:      "internal server error",
			expenseID: "exp-123",
			requestBody: dtos.UpdateExpenseStatusRequest{
				Status: domain.Approved,
			},
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("UpdateExpenseStatus", context.Background(), "exp-123", domain.Approved).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ExpenseService)
			tt.setupMock(mockService)

			handler := NewExpenseHandler(mockService)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/expenses/"+tt.expenseID+"/status", bytes.NewBuffer(body))
			req.SetPathValue("id", tt.expenseID)
			w := httptest.NewRecorder()

			handler.HandleUpdateExpenseStatus(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err = json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}

func Test_handlers_ExpenseHandler_HandleGetExpenseSummary(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mockservice.ExpenseService)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful get expense summary",
			setupMock: func(m *mockservice.ExpenseService) {
				summary := domain.ExpenseSummary{
					TotalExpenses:     5000.00,
					PendingExpense:    2000.00,
					ReimbursedExpense: 2500.00,
					RejectedExpense:   500.00,
				}
				m.On("GetExpenseSummary", context.Background()).Return(summary, nil)
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "expense summary fetched successfully",
		},
		{
			name: "unauthorized",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseSummary", context.Background()).
					Return(domain.ExpenseSummary{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", errors.New("unauthorized")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name: "forbidden",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseSummary", context.Background()).
					Return(domain.ExpenseSummary{}, apperr.NewAppError(apperr.ErrForbidden, "forbidden", errors.New("forbidden")))
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name: "internal server error",
			setupMock: func(m *mockservice.ExpenseService) {
				m.On("GetExpenseSummary", context.Background()).
					Return(domain.ExpenseSummary{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockservice.ExpenseService)
			tt.setupMock(mockService)

			handler := NewExpenseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/expenses/summary", nil)
			w := httptest.NewRecorder()

			handler.HandleGetExpenseSummary(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dtos.BaseResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, response.Message)

			mockService.AssertExpectations(t)
		})
	}
}
