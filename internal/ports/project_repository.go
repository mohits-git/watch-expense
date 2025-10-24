package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type ProjectRepository interface {
  SaveProject(ctx context.Context, project domain.Project) (string, error)
  UpdateProject(ctx context.Context, project domain.Project) error
  FindProjectById(ctx context.Context, projectId string) (domain.Project, error)
  FindAllProjects(ctx context.Context) ([]domain.Project, error)
}
