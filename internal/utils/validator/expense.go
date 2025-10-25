package validator

import "github.com/mohits-git/watch-expense/internal/domain"

func ValidateExpenseCreation(expense domain.Expense) bool {
	if expense.Amount <= 0 ||
		expense.UserID == "" ||
		!ValidateUUID(expense.UserID) ||
		expense.Purpose == "" ||
		expense.Status != domain.Pending {
		return false
	}
	return true
}

func ValidateExpenseUpdate(expense domain.Expense) bool {
	if !ValidateUUID(expense.ID) ||
		expense.Amount <= 0 ||
		expense.UserID == "" ||
		!ValidateUUID(expense.UserID) ||
		expense.Purpose == "" ||
		expense.Status != domain.Pending {
		return false
	}
	return true
}

func ValidateExpenseStatus(status domain.RequestStatus) bool {
	switch status {
	case domain.Pending, domain.Approved, domain.Rejected, domain.Reviewed:
		return true
	default:
		return false
	}
}
