package validator

import "github.com/mohits-git/watch-expense/internal/domain"

func ValidateProjectCreation(project domain.Project) bool {
	if project.Name == "" ||
		project.Budget < 0 {
		return false
	}

	if project.DepartmentID != "" && !ValidateUUID(project.DepartmentID) {
		return false
	}

	return true
}

func ValidateProjectUpdate(project domain.Project) bool {
	if !ValidateUUID(project.ID) ||
		project.Name == "" ||
		project.Budget < 0 {
		return false
	}

	if project.DepartmentID != "" && !ValidateUUID(project.DepartmentID) {
		return false
	}
	return true
}
