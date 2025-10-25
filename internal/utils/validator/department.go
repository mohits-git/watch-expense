package validator

import "github.com/mohits-git/watch-expense/internal/domain"

func ValidateDepartmentCreation(department domain.Department) bool {
	if department.Name == "" ||
		department.Budget < 0 {
		return false
	}
	return true
}

func ValidateDepartmentUpdate(department domain.Department) bool {
	if !ValidateUUID(department.ID) ||
		department.Name == "" ||
		department.Budget < 0 {
		return false
	}
	return true
}
