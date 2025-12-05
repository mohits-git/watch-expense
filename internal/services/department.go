package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	"github.com/mohits-git/watch-expense/internal/utils/validator"
)

type DepartmentService interface {
	CreateDepartment(ctx context.Context, department domain.Department) (string, error)
	GetDepartmentByID(ctx context.Context, departmentID string) (domain.Department, error)
	UpdateDepartment(ctx context.Context, department domain.Department) error
	GetAllDepartments(ctx context.Context) ([]domain.Department, error)
}

type departmentService struct {
	departmentRepo ports.DepartmentRepository
}

func NewDepartmentService(departmentRepo ports.DepartmentRepository) DepartmentService {
	return &departmentService{
		departmentRepo: departmentRepo,
	}
}

func (s *departmentService) CreateDepartment(ctx context.Context, department domain.Department) (string, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		return "", apperr.NewAppError(apperr.ErrForbidden, "only admin can create departments", nil)
	}

	if !validator.ValidateDepartmentCreation(department) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid department data", nil)
	}

	department.ID = uuid.New().String()
	department.CreatedAt = time.Now().UnixMilli()
	department.UpdatedAt = time.Now().UnixMilli()

	return s.departmentRepo.SaveDepartment(ctx, department)
}

func (s *departmentService) GetDepartmentByID(ctx context.Context, departmentID string) (domain.Department, error) {
	if !validator.ValidateUUID(departmentID) {
		return domain.Department{}, apperr.NewAppError(apperr.ErrInvalid, "invalid department ID", nil)
	}

	_, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.Department{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	return s.departmentRepo.FindDepartmentById(ctx, departmentID)
}

func (s *departmentService) UpdateDepartment(ctx context.Context, department domain.Department) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrForbidden, "only admin can update departments", nil)
	}

	if !validator.ValidateUUID(department.ID) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid department ID", nil)
	}

	if !validator.ValidateDepartmentUpdate(department) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid department data", nil)
	}

	existingDepartment, err := s.departmentRepo.FindDepartmentById(ctx, department.ID)
	if err != nil {
		return err
	}

	department.UpdatedAt = time.Now().UnixMilli()
	department.CreatedAt = existingDepartment.CreatedAt

	return s.departmentRepo.UpdateDepartment(ctx, department)
}

func (s *departmentService) GetAllDepartments(ctx context.Context) ([]domain.Department, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return nil, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}
	if claims.Role != domain.Admin {
		return nil, apperr.NewAppError(apperr.ErrForbidden, "only admin can view departments", nil)
	}

	return s.departmentRepo.FindAllDepartments(ctx)
}
