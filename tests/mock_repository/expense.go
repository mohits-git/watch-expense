package mockrepository

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/stretchr/testify/mock"
)

type ExpenseRepository struct {
	mock.Mock
}

func NewMockExpenseRepository() ports.ExpenseRepository {
	return &ExpenseRepository{}
}

func (e *ExpenseRepository) SaveExpense(ctx context.Context, expense domain.Expense) (string, error) {
	args := e.Called(ctx, expense)
	return args.String(0), args.Error(1)
}

func (e *ExpenseRepository) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	args := e.Called(ctx, expense)
	return args.Error(0)
}

func (e *ExpenseRepository) FindExpenseById(ctx context.Context, expenseId string) (domain.Expense, error) {
	args := e.Called(ctx, expenseId)
	return args.Get(0).(domain.Expense), args.Error(1)
}

func (e *ExpenseRepository) FindExpensesByUserId(ctx context.Context, userId string) ([]domain.Expense, error) {
	args := e.Called(ctx, userId)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (e *ExpenseRepository) FindAllExpenses(ctx context.Context) ([]domain.Expense, error) {
	args := e.Called(ctx)
	return args.Get(0).([]domain.Expense), args.Error(1)
}
