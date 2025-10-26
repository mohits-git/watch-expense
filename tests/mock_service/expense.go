package mockservice

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/mock"
)

type ExpenseService struct {
	mock.Mock
}

func NewMockExpenseService() *ExpenseService {
	return &ExpenseService{}
}

func (e *ExpenseService) GetExpenseByID(ctx context.Context, expenseID string) (domain.Expense, error) {
	args := e.Called(ctx, expenseID)
	return args.Get(0).(domain.Expense), args.Error(1)
}

func (e *ExpenseService) CreateExpense(ctx context.Context, expense domain.Expense) (string, error) {
	args := e.Called(ctx, expense)
	return args.String(0), args.Error(1)
}

func (e *ExpenseService) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	args := e.Called(ctx, expense)
	return args.Error(0)
}

func (e *ExpenseService) GetAllExpenses(ctx context.Context, filterOptions domain.ExpensesFilterOptions) ([]domain.Expense, int, error) {
	args := e.Called(ctx, filterOptions)
	return args.Get(0).([]domain.Expense), args.Int(1), args.Error(2)
}

func (e *ExpenseService) UpdateExpenseStatus(ctx context.Context, expenseID string, status domain.RequestStatus) error {
	args := e.Called(ctx, expenseID, status)
	return args.Error(0)
}

func (e *ExpenseService) GetExpenseSummary(ctx context.Context) (domain.ExpenseSummary, error) {
	args := e.Called(ctx)
	return args.Get(0).(domain.ExpenseSummary), args.Error(1)
}
