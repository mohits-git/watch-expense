package mockrepository

import (
  "context"
  "github.com/mohits-git/watch-expense/internal/domain"
  "github.com/stretchr/testify/mock"
)

type AdvanceRepository struct {
  mock.Mock
}

func NewMockAdvanceRepository() *AdvanceRepository {
  return &AdvanceRepository{}
}

func (a *AdvanceRepository) SaveAdvance(ctx context.Context, advance domain.Advance) (string, error) {
  args := a.Called(ctx, advance)
  return args.String(0), args.Error(1)
}

func (a *AdvanceRepository) UpdateAdvance(ctx context.Context, advance domain.Advance) error {
  args := a.Called(ctx, advance)
  return args.Error(0)
}

func (a *AdvanceRepository) FindAdvanceById(ctx context.Context, advanceId string) (domain.Advance, error) {
  args := a.Called(ctx, advanceId)
  return args.Get(0).(domain.Advance), args.Error(1)
}

func (a *AdvanceRepository) FindAdvancesByUserId(ctx context.Context, userId string) ([]domain.Advance, error) {
  args := a.Called(ctx, userId)
  return args.Get(0).([]domain.Advance), args.Error(1)
}

func (a *AdvanceRepository) FindAllAdvances(ctx context.Context) ([]domain.Advance, error) {
  args := a.Called(ctx)
  return args.Get(0).([]domain.Advance), args.Error(1)
}
