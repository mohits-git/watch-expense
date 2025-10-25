package mockservice

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/mock"
)

type ProjectService struct {
	mock.Mock
}

func NewMockProjectService() *ProjectService {
	return &ProjectService{}
}

func (p *ProjectService) CreateProject(ctx context.Context, project domain.Project) (string, error) {
	args := p.Called(ctx, project)
	return args.String(0), args.Error(1)
}

func (p *ProjectService) GetProjectByID(ctx context.Context, projectID string) (domain.Project, error) {
	args := p.Called(ctx, projectID)
	return args.Get(0).(domain.Project), args.Error(1)
}

func (p *ProjectService) UpdateProject(ctx context.Context, project domain.Project) error {
	args := p.Called(ctx, project)
	return args.Error(0)
}

func (p *ProjectService) GetAllProjects(ctx context.Context) ([]domain.Project, error) {
	args := p.Called(ctx)
	return args.Get(0).([]domain.Project), args.Error(1)
}
