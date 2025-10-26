package mockservice

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/mock"
)

type AdvanceService struct {
	mock.Mock
}

func NewMockAdvanceService() *AdvanceService {
	return &AdvanceService{}
}

func (a *AdvanceService) CreateAdvance(ctx context.Context, advance domain.Advance) (string, error) {
	args := a.Called(ctx, advance)
	return args.String(0), args.Error(1)
}

func (a *AdvanceService) GetAdvanceByID(ctx context.Context, advanceID string) (domain.Advance, error) {
	args := a.Called(ctx, advanceID)
	return args.Get(0).(domain.Advance), args.Error(1)
}

func (a *AdvanceService) UpdateAdvance(ctx context.Context, advance domain.Advance) error {
	args := a.Called(ctx, advance)
	return args.Error(0)
}

func (a *AdvanceService) UpdateAdvanceStatus(ctx context.Context, advanceID string, status domain.RequestStatus) error {
	args := a.Called(ctx, advanceID, status)
	return args.Error(0)
}

func (a *AdvanceService) GetAllAdvances(ctx context.Context, filterOptions domain.AdvancesFilterOptions) ([]domain.Advance, int, error) {
	args := a.Called(ctx, filterOptions)
	return args.Get(0).([]domain.Advance), args.Int(1), args.Error(2)
}

func (a *AdvanceService) GetAdvanceSummary(ctx context.Context) (domain.AdvanceSummary, error) {
	args := a.Called(ctx)
	return args.Get(0).(domain.AdvanceSummary), args.Error(1)
}
