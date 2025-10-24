package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type AdvanceRepository interface {
  SaveAdvance(ctx context.Context, advance domain.Advance) (string, error)
  UpdateAdvance(ctx context.Context, advance domain.Advance) error
  FindAdvanceById(ctx context.Context, advanceId string) (domain.Advance, error)
  FindAdvancesByUserId(ctx context.Context, userId string) ([]domain.Advance, error)
  FindAllAdvances(ctx context.Context) ([]domain.Advance, error)
}
