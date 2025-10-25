package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	"github.com/mohits-git/watch-expense/internal/utils/validator"
)

type ExpenseService interface {
	GetExpenseByID(ctx context.Context, expenseID string) (domain.Expense, error)
	CreateExpense(ctx context.Context, expense domain.Expense) (string, error)
	UpdateExpense(ctx context.Context, expense domain.Expense) error
	GetAllExpenses(ctx context.Context, filterOptions domain.ExpensesFilterOptions) ([]domain.Expense, int, error)
	UpdateExpenseStatus(ctx context.Context, expenseID string, status domain.RequestStatus) error
}

type expenseService struct {
	expenseRepo ports.ExpenseRepository
}

func NewExpenseService(expenseRepo ports.ExpenseRepository) ExpenseService {
	return &expenseService{
		expenseRepo: expenseRepo,
	}
}

func (s *expenseService) CreateExpense(ctx context.Context, expense domain.Expense) (string, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Employee {
		return "", apperr.NewAppError(apperr.ErrForbidden, "only employees can create expenses", nil)
	}

	expense.UserID = claims.UserID

	expense.Status = domain.Pending

	if !validator.ValidateExpenseCreation(expense) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid expense data", nil)
	}

	expense.ID = uuid.New().String()
	expense.CreatedAt = time.Now().Unix()
	expense.UpdatedAt = time.Now().Unix()

	return s.expenseRepo.SaveExpense(ctx, expense)
}

func (s *expenseService) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Employee {
		return apperr.NewAppError(apperr.ErrForbidden, "only employees can update expenses", nil)
	}

	if !validator.ValidateUUID(expense.ID) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid expense ID", nil)
	}

	existingExpense, err := s.expenseRepo.FindExpenseById(ctx, expense.ID)
	if err != nil {
		return err
	}

	if existingExpense.UserID != claims.UserID {
		return apperr.NewAppError(apperr.ErrForbidden, "you can only update your own expenses", nil)
	}

	if existingExpense.Status != domain.Pending {
		return apperr.NewAppError(apperr.ErrForbidden, "cannot update expense that is not pending", nil)
	}

	expense.Status = domain.Pending
	expense.UserID = claims.UserID

	if !validator.ValidateExpenseUpdate(expense) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid expense data", nil)
	}

	expense.UpdatedAt = time.Now().Unix()
	expense.CreatedAt = existingExpense.CreatedAt
	expense.ApprovedAt = existingExpense.ApprovedAt
	expense.ApprovedBy = existingExpense.ApprovedBy
	expense.ReviewedAt = existingExpense.ReviewedAt
	expense.ReviewedBy = existingExpense.ReviewedBy

	return s.expenseRepo.UpdateExpense(ctx, expense)
}

func (s *expenseService) UpdateExpenseStatus(ctx context.Context, expenseID string, status domain.RequestStatus) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrForbidden, "only admin can update expense status", nil)
	}

	if !validator.ValidateUUID(expenseID) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid expense ID", nil)
	}

	if !validator.ValidateExpenseStatus(status) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid status", nil)
	}

	expense, err := s.expenseRepo.FindExpenseById(ctx, expenseID)
	if err != nil {
		return err
	}

	reviewerID := claims.UserID

	expense.Status = status

	if status == domain.Approved {
		expense.ApprovedBy = reviewerID
		expense.ApprovedAt = time.Now().Unix()
	}

	if status == domain.Reviewed {
		expense.ReviewedBy = reviewerID
		expense.ReviewedAt = time.Now().Unix()
	}

	expense.UpdatedAt = time.Now().Unix()

	return s.expenseRepo.UpdateExpense(ctx, expense)
}

func (s *expenseService) GetExpenseByID(ctx context.Context, expenseID string) (domain.Expense, error) {
	if !validator.ValidateUUID(expenseID) {
		return domain.Expense{}, apperr.NewAppError(apperr.ErrInvalid, "invalid expense ID", nil)
	}

	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.Expense{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	expense, err := s.expenseRepo.FindExpenseById(ctx, expenseID)
	if err != nil {
		return domain.Expense{}, err
	}

	if claims.Role != domain.Admin && expense.UserID != claims.UserID {
		return domain.Expense{}, apperr.NewAppError(apperr.ErrForbidden, "access denied", nil)
	}

	return expense, nil
}

func (s *expenseService) GetAllExpenses(ctx context.Context, filterOptions domain.ExpensesFilterOptions) ([]domain.Expense, int, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return nil, 0, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		filterOptions.UserID = claims.UserID
	}

	return s.expenseRepo.FindAllExpenses(ctx, filterOptions)
}
