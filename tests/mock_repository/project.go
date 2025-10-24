package mockrepository

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/stretchr/testify/mock"
)

type ProjectRepository struct {
  mock.Mock
}

func NewMockProjectRepository() ports.ProjectRepository {
  return &ProjectRepository{}
}

func (p *ProjectRepository) SaveProject(ctx context.Context, project domain.Project) (string, error) {
  args := p.Called(ctx, project)
  return args.String(0), args.Error(1)
}

func (p *ProjectRepository) UpdateProject(ctx context.Context, project domain.Project) error {
  args := p.Called(ctx, project)
  return args.Error(0)
}

func (p *ProjectRepository) FindProjectById(ctx context.Context, projectId string) (domain.Project, error) {
  args := p.Called(ctx, projectId)
  return args.Get(0).(domain.Project), args.Error(1)
}

func (p *ProjectRepository) FindAllProjects(ctx context.Context) ([]domain.Project, error) {
  args := p.Called(ctx)
  return args.Get(0).([]domain.Project), args.Error(1)
}
