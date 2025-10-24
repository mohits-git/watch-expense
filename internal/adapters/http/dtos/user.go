package dtos

import "github.com/mohits-git/watch-expense/internal/domain"

type User struct {
	ID           string          `json:"id"`
	EmployeeId   string          `json:"employeeId"`
	Name         string          `json:"name"`
	Password     string          `json:"-"`
	Email        string          `json:"email"`
	Role         domain.UserRole `json:"role"`
	ProjectID    string          `json:"projectId"`
	DepartmentID string          `json:"departmentId"`
	CreatedAt    int64           `json:"createdAt"`
	UpdatedAt    int64           `json:"updatedAt"`
}
