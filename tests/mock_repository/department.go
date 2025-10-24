package mockrepository

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/stretchr/testify/mock"
)

type DepartmentRepository struct {
  mock.Mock
}

func NewMockDepartmentRepository() ports.DepartmentRepository {
  return &DepartmentRepository{}
}

func (d *DepartmentRepository) SaveDepartment(ctx context.Context, department domain.Department) (string, error) {
  args := d.Called(ctx, department)
  return args.String(0), args.Error(1)
}

func (d *DepartmentRepository) UpdateDepartment(ctx context.Context, department domain.Department) error {
  args := d.Called(ctx, department)
  return args.Error(0)
}

func (d *DepartmentRepository) FindDepartmentById(ctx context.Context, departmentId string) (domain.Department, error) {
  args := d.Called(ctx, departmentId)
  return args.Get(0).(domain.Department), args.Error(1)
}

func (d *DepartmentRepository) FindAllDepartments(ctx context.Context) ([]domain.Department, error) {
  args := d.Called(ctx)
  return args.Get(0).([]domain.Department), args.Error(1)
}
