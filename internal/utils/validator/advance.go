package validator

import "github.com/mohits-git/watch-expense/internal/domain"

func ValidateAdvanceCreation(advance domain.Advance) bool {
	if advance.Amount <= 0 ||
		advance.UserID == "" ||
		!ValidateUUID(advance.UserID) ||
		advance.Purpose == "" ||
		advance.Status != domain.Pending {
		return false
	}
	return true
}

func ValidateAdvanceUpdate(advance domain.Advance) bool {
	if !ValidateUUID(advance.ID) ||
		advance.Amount <= 0 ||
		advance.UserID == "" ||
		!ValidateUUID(advance.UserID) ||
		advance.Purpose == "" ||
		advance.Status != domain.Pending {
		return false
	}
	return true
}
