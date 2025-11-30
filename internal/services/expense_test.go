package services

import (
	"context"
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockrepository "github.com/mohits-git/watch-expense/tests/mock_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_services_NewExpenseService(t *testing.T) {
	expenseRepo := mockrepository.NewMockExpenseRepository()
	expenseService := NewExpenseService(expenseRepo)
	require.NotNil(t, expenseService, "NewExpenseService() returned nil")
}

func Test_services_ExpenseService_CreateExpense(t *testing.T) {
	employeeID := "650e8400-e29b-41d4-a716-446655440000"
	type args struct {
		ctx     context.Context
		expense domain.Expense
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ExpenseRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "create expense successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					Amount:      1000.00,
					Description: "Test expense",
					Purpose:     "Business travel",
					Bills: []domain.Bill{
						{Amount: 500.00, Description: "Hotel", AttachmentURL: "url1"},
						{Amount: 500.00, Description: "Transport", AttachmentURL: "url2"},
					},
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("SaveExpense", mock.Anything, mock.AnythingOfType("domain.Expense")).Return("new-expense-id", nil)
				return expenseRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				expense: domain.Expense{
					Amount:      1000.00,
					Description: "Test expense",
					Purpose:     "Business travel",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Admin,
				}),
				expense: domain.Expense{
					Amount:      1000.00,
					Description: "Test expense",
					Purpose:     "Business travel",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid expense data - zero amount",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					Amount:      0,
					Description: "Test expense",
					Purpose:     "Business travel",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid expense data - empty purpose",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					Amount:      1000.00,
					Description: "Business trip",
					Purpose:     "",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := tt.getMockRepo()
			expenseService := NewExpenseService(expenseRepo)

			result, err := expenseService.CreateExpense(tt.args.ctx, tt.args.expense)

			if tt.wantErr {
				assert.Error(t, err, "CreateExpense() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "CreateExpense() should not return error")
				assert.NotEmpty(t, result, "CreateExpense() should return expense ID")
			}

			expenseRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ExpenseService_UpdateExpense(t *testing.T) {
	validExpenseID := "550e8400-e29b-41d4-a716-446655440000"
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx     context.Context
		expense domain.Expense
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ExpenseRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update expense successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					ID:          validExpenseID,
					Amount:      1500.00,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:        validExpenseID,
					UserID:    employeeID,
					Status:    domain.Pending,
					CreatedAt: 1234567890,
				}, nil)
				expenseRepo.On("UpdateExpense", mock.Anything, mock.AnythingOfType("domain.Expense")).Return(nil)
				return expenseRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				expense: domain.Expense{
					ID:          validExpenseID,
					Amount:      1500.00,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				expense: domain.Expense{
					ID:          validExpenseID,
					Amount:      1500.00,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid expense ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					ID:          "invalid-id",
					Amount:      1500.00,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "forbidden - updating another user's expense",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					ID:          validExpenseID,
					Amount:      1500.00,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					UserID: "other-employee-id",
					Status: domain.Pending,
				}, nil)
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "forbidden - expense is not pending",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					ID:          validExpenseID,
					Amount:      1500.00,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					UserID: employeeID,
					Status: domain.Approved,
				}, nil)
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid expense data - zero amount",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expense: domain.Expense{
					ID:          validExpenseID,
					Amount:      0,
					Description: "Updated expense",
					Purpose:     "Updated purpose",
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					UserID: employeeID,
					Status: domain.Pending,
				}, nil)
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := tt.getMockRepo()
			expenseService := NewExpenseService(expenseRepo)

			err := expenseService.UpdateExpense(tt.args.ctx, tt.args.expense)

			if tt.wantErr {
				assert.Error(t, err, "UpdateExpense() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateExpense() should not return error")
			}

			expenseRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ExpenseService_UpdateExpenseStatus(t *testing.T) {
	validExpenseID := "550e8400-e29b-41d4-a716-446655440000"
	adminID := "750e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx       context.Context
		expenseID string
		status    domain.RequestStatus
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ExpenseRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update expense status to approved successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				expenseID: validExpenseID,
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					Status: domain.Pending,
				}, nil)
				expenseRepo.On("UpdateExpense", mock.Anything, mock.AnythingOfType("domain.Expense")).Return(nil)
				return expenseRepo
			},
			wantErr: false,
		},
		{
			name: "update expense status to rejected successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				expenseID: validExpenseID,
				status:    domain.Rejected,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					Status: domain.Pending,
				}, nil)
				expenseRepo.On("UpdateExpense", mock.Anything, mock.AnythingOfType("domain.Expense")).Return(nil)
				return expenseRepo
			},
			wantErr: false,
		},
		{
			name: "update expense status to reviewed successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				expenseID: validExpenseID,
				status:    domain.Reviewed,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					Status: domain.Approved,
				}, nil)
				expenseRepo.On("UpdateExpense", mock.Anything, mock.AnythingOfType("domain.Expense")).Return(nil)
				return expenseRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:       context.Background(),
				expenseID: validExpenseID,
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is employee",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "employee-id",
					Role:   domain.Employee,
				}),
				expenseID: validExpenseID,
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid expense ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				expenseID: "invalid-id",
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid status - pending not allowed",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				expenseID: validExpenseID,
				status:    domain.Pending,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := tt.getMockRepo()
			expenseService := NewExpenseService(expenseRepo)

			err := expenseService.UpdateExpenseStatus(tt.args.ctx, tt.args.expenseID, tt.args.status)

			if tt.wantErr {
				assert.Error(t, err, "UpdateExpenseStatus() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateExpenseStatus() should not return error")
			}

			expenseRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ExpenseService_GetExpenseByID(t *testing.T) {
	validExpenseID := "550e8400-e29b-41d4-a716-446655440000"
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx       context.Context
		expenseID string
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ExpenseRepository
		want        domain.Expense
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get expense by ID successfully - admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				expenseID: validExpenseID,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:          validExpenseID,
					UserID:      employeeID,
					Amount:      1000.00,
					Description: "Test expense",
					Status:      domain.Pending,
				}, nil)
				return expenseRepo
			},
			want: domain.Expense{
				ID:          validExpenseID,
				UserID:      employeeID,
				Amount:      1000.00,
				Description: "Test expense",
				Status:      domain.Pending,
			},
			wantErr: false,
		},
		{
			name: "get expense by ID successfully - own expense",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				expenseID: validExpenseID,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:          validExpenseID,
					UserID:      employeeID,
					Amount:      1000.00,
					Description: "Test expense",
					Status:      domain.Pending,
				}, nil)
				return expenseRepo
			},
			want: domain.Expense{
				ID:          validExpenseID,
				UserID:      employeeID,
				Amount:      1000.00,
				Description: "Test expense",
				Status:      domain.Pending,
			},
			wantErr: false,
		},
		{
			name: "invalid expense ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				expenseID: "invalid-id",
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:       context.Background(),
				expenseID: validExpenseID,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - accessing another user's expense",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "other-employee-id",
					Role:   domain.Employee,
				}),
				expenseID: validExpenseID,
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("FindExpenseById", mock.Anything, validExpenseID).Return(domain.Expense{
					ID:     validExpenseID,
					UserID: employeeID,
				}, nil)
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := tt.getMockRepo()
			expenseService := NewExpenseService(expenseRepo)

			result, err := expenseService.GetExpenseByID(tt.args.ctx, tt.args.expenseID)

			if tt.wantErr {
				assert.Error(t, err, "GetExpenseByID() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetExpenseByID() should not return error")
				assert.Equal(t, tt.want, result, "GetExpenseByID() returned incorrect expense")
			}

			expenseRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ExpenseService_GetAllExpenses(t *testing.T) {
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx           context.Context
		filterOptions domain.ExpensesFilterOptions
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ExpenseRepository
		wantList    []domain.Expense
		wantTotal   int
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get all expenses successfully - admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				filterOptions: domain.ExpensesFilterOptions{
					Page:  1,
					Limit: 10,
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenses := []domain.Expense{
					{ID: "expense-1", UserID: employeeID, Amount: 1000.00},
					{ID: "expense-2", UserID: "other-user", Amount: 2000.00},
				}
				expenseRepo.On("FindAllExpenses", mock.Anything, mock.AnythingOfType("domain.ExpensesFilterOptions")).
					Return(expenses, 2, nil)
				return expenseRepo
			},
			wantList: []domain.Expense{
				{ID: "expense-1", UserID: employeeID, Amount: 1000.00},
				{ID: "expense-2", UserID: "other-user", Amount: 2000.00},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "get expenses for employee - filtered by userID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				filterOptions: domain.ExpensesFilterOptions{
					Page:  1,
					Limit: 10,
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenses := []domain.Expense{
					{ID: "expense-1", UserID: employeeID, Amount: 1000.00},
				}
				// Employee requests should have their UserID automatically set
				expenseRepo.On("FindAllExpenses", mock.Anything, mock.MatchedBy(func(filter domain.ExpensesFilterOptions) bool {
					return filter.UserID == employeeID
				})).Return(expenses, 1, nil)
				return expenseRepo
			},
			wantList: []domain.Expense{
				{ID: "expense-1", UserID: employeeID, Amount: 1000.00},
			},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				filterOptions: domain.ExpensesFilterOptions{
					Page:  1,
					Limit: 10,
				},
			},
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := tt.getMockRepo()
			expenseService := NewExpenseService(expenseRepo)

			resultList, resultTotal, err := expenseService.GetAllExpenses(tt.args.ctx, tt.args.filterOptions)

			if tt.wantErr {
				assert.Error(t, err, "GetAllExpenses() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAllExpenses() should not return error")
				assert.Equal(t, tt.wantList, resultList, "GetAllExpenses() returned incorrect expenses")
				assert.Equal(t, tt.wantTotal, resultTotal, "GetAllExpenses() returned incorrect total")
			}

			expenseRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ExpenseService_GetExpenseSummary(t *testing.T) {
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name        string
		ctx         context.Context
		getMockRepo func() *mockrepository.ExpenseRepository
		want        domain.ExpenseSummary
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get expense summary successfully - admin",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "admin-id",
				Role:   domain.Admin,
			}),
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, "", domain.RequestStatus("")).Return(10000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, "", domain.Pending).Return(3000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, "", domain.Approved).Return(5000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, "", domain.Rejected).Return(2000.00, nil)
				return expenseRepo
			},
			want: domain.ExpenseSummary{
				TotalExpenses:     10000.00,
				PendingExpense:    3000.00,
				ReimbursedExpense: 5000.00,
				RejectedExpense:   2000.00,
			},
			wantErr: false,
		},
		{
			name: "get expense summary successfully - employee",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: employeeID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.RequestStatus("")).Return(5000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.Pending).Return(1000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.Approved).Return(3000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.Rejected).Return(1000.00, nil)
				return expenseRepo
			},
			want: domain.ExpenseSummary{
				TotalExpenses:     5000.00,
				PendingExpense:    1000.00,
				ReimbursedExpense: 3000.00,
				RejectedExpense:   1000.00,
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			ctx:  context.Background(),
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "error from repository - total expenses",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: employeeID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() *mockrepository.ExpenseRepository {
				expenseRepo := mockrepository.NewMockExpenseRepository()
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.RequestStatus("")).
					Return(0.0, apperr.NewAppError(apperr.ErrInternal, "database error", nil))
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.Pending).Return(1000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.Approved).Return(3000.00, nil)
				expenseRepo.On("GetExpenseSumByStatus", mock.Anything, employeeID, domain.Rejected).Return(1000.00, nil)
				return expenseRepo
			},
			wantErr: true,
			errCode: apperr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := tt.getMockRepo()
			expenseService := NewExpenseService(expenseRepo)

			result, err := expenseService.GetExpenseSummary(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetExpenseSummary() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetExpenseSummary() should not return error")
				assert.Equal(t, tt.want, result, "GetExpenseSummary() returned incorrect summary")
			}

			expenseRepo.AssertExpectations(t)
		})
	}
}
