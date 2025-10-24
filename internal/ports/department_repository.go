package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type DepartmentRepository interface {
  SaveDepartment(ctx context.Context, department domain.Department) (string, error)
  UpdateDepartment(ctx context.Context, department domain.Department) error
  FindDepartmentById(ctx context.Context, departmentId string) (domain.Department, error)
  FindAllDepartments(ctx context.Context) ([]domain.Department, error)
}
