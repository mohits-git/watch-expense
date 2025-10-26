package validator

import "github.com/mohits-git/watch-expense/internal/domain"

func ValidateUserCreation(user domain.User) bool {
	if !ValidateEmail(user.Email) ||
  // !ValidatePassword(user.Password) || // TODO:
		!ValidateUserRole(user.Role) ||
		user.Name == "" {
		return false
	}
	return true
}

func ValidateUserUpdate(user domain.User) bool {
	if !ValidateUUID(user.ID) ||
		(user.Email != "" && !ValidateEmail(user.Email)) ||
		(user.Password != "" && !ValidatePassword(user.Password)) ||
		(user.Role != "" && !ValidateUserRole(user.Role)) ||
		(user.DepartmentID != "" && !ValidateUUID(user.DepartmentID)) ||
		(user.ProjectID != "" && !ValidateUUID(user.ProjectID)) ||
		user.Name == "" {
		return false
	}
	return true
}

func ValidateUserRole(role domain.UserRole) bool {
	switch role {
	case domain.Admin, domain.Employee, domain.Manager:
		return true
	default:
		return false
	}
}
