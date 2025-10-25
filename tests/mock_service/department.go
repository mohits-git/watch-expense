package mockservice

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/mock"
)

type DepartmentService struct {
	mock.Mock
}

func NewMockDepartmentService() *DepartmentService {
	return &DepartmentService{}
}

func (d *DepartmentService) CreateDepartment(ctx context.Context, department domain.Department) (string, error) {
	args := d.Called(ctx, department)
	return args.String(0), args.Error(1)
}

func (d *DepartmentService) GetDepartmentByID(ctx context.Context, departmentID string) (domain.Department, error) {
	args := d.Called(ctx, departmentID)
	return args.Get(0).(domain.Department), args.Error(1)
}

func (d *DepartmentService) UpdateDepartment(ctx context.Context, department domain.Department) error {
	args := d.Called(ctx, department)
	return args.Error(0)
}

func (d *DepartmentService) GetAllDepartments(ctx context.Context) ([]domain.Department, error) {
	args := d.Called(ctx)
	return args.Get(0).([]domain.Department), args.Error(1)
}
