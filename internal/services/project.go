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

type ProjectService interface {
	CreateProject(ctx context.Context, project domain.Project) (string, error)
	GetProjectByID(ctx context.Context, projectID string) (domain.Project, error)
	UpdateProject(ctx context.Context, project domain.Project) error
	GetAllProjects(ctx context.Context) ([]domain.Project, error)
}

type projectService struct {
	projectRepo ports.ProjectRepository
}

func NewProjectService(projectRepo ports.ProjectRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

func (s *projectService) CreateProject(ctx context.Context, project domain.Project) (string, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		return "", apperr.NewAppError(apperr.ErrForbidden, "only admin can create projects", nil)
	}

	if !validator.ValidateProjectCreation(project) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid project data", nil)
	}

	if project.DepartmentID != "" && !validator.ValidateUUID(project.DepartmentID) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid department ID", nil)
	}

	project.ID = uuid.New().String()
	project.CreatedAt = time.Now().UnixMilli()
	project.UpdatedAt = time.Now().UnixMilli()

	return s.projectRepo.SaveProject(ctx, project)
}

func (s *projectService) GetProjectByID(ctx context.Context, projectID string) (domain.Project, error) {
	if !validator.ValidateUUID(projectID) {
		return domain.Project{}, apperr.NewAppError(apperr.ErrInvalid, "invalid project ID", nil)
	}

	_, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.Project{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	return s.projectRepo.FindProjectById(ctx, projectID)
}

func (s *projectService) UpdateProject(ctx context.Context, project domain.Project) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrForbidden, "only admin can update projects", nil)
	}

	if !validator.ValidateUUID(project.ID) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid project ID", nil)
	}

	if !validator.ValidateProjectUpdate(project) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid project data", nil)
	}

	if project.DepartmentID != "" && !validator.ValidateUUID(project.DepartmentID) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid department ID", nil)
	}

	existingProject, err := s.projectRepo.FindProjectById(ctx, project.ID)
	if err != nil {
		return err
	}

	project.UpdatedAt = time.Now().UnixMilli()
	project.CreatedAt = existingProject.CreatedAt

	return s.projectRepo.UpdateProject(ctx, project)
}

func (s *projectService) GetAllProjects(ctx context.Context) ([]domain.Project, error) {
	_, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return nil, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	return s.projectRepo.FindAllProjects(ctx)
}
