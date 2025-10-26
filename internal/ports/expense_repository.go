package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type ExpenseRepository interface {
	SaveExpense(ctx context.Context, expense domain.Expense) (string, error)
	UpdateExpense(ctx context.Context, expense domain.Expense) error
	FindExpenseById(ctx context.Context, expenseId string) (domain.Expense, error)
	FindAllExpenses(ctx context.Context, filterOptions domain.ExpensesFilterOptions) ([]domain.Expense, int, error)
	GetExpenseSumByStatus(ctx context.Context, userID string, status domain.RequestStatus) (float64, error)
}
