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

func ToUserDTO(user domain.User) User {
	return User{
		ID:           user.ID,
		EmployeeId:   user.EmployeeId,
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		ProjectID:    user.ProjectID,
		DepartmentID: user.DepartmentID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func ToUserDomain(dto User) domain.User {
	return domain.User{
		EmployeeId:   dto.EmployeeId,
		Name:         dto.Name,
		Password:     dto.Password,
		Email:        dto.Email,
		Role:         dto.Role,
		ProjectID:    dto.ProjectID,
		DepartmentID: dto.DepartmentID,
	}
}

type CreateUserRequest struct {
	EmployeeId   string          `json:"employeeId"`
	Name         string          `json:"name"`
	Password     string          `json:"password"`
	Email        string          `json:"email"`
	Role         domain.UserRole `json:"role"`
	ProjectID    string          `json:"projectId"`
	DepartmentID string          `json:"departmentId"`
}

type CreateUserResponse struct {
	ID string `json:"id"`
}

type UpdateUserRequest struct {
	EmployeeId   string          `json:"employeeId"`
	Name         string          `json:"name"`
	Password     string          `json:"password"`
	Email        string          `json:"email"`
	Role         domain.UserRole `json:"role"`
	ProjectID    string          `json:"projectId"`
	DepartmentID string          `json:"departmentId"`
}

type GetUserBudgetResponse struct {
	Budget float64 `json:"budget"`
}
