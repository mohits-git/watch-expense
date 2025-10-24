package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type ExpenseRepository interface {
  SaveExpense(ctx context.Context, expense domain.Expense) (string, error)
  UpdateExpense(ctx context.Context, expense domain.Expense) error
  FindExpenseById(ctx context.Context, expenseId string) (domain.Expense, error)
  FindExpensesByUserId(ctx context.Context, userId string) ([]domain.Expense, error)
  FindAllExpenses(ctx context.Context) ([]domain.Expense, error)
}
