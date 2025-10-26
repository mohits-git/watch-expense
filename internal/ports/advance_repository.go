package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type AdvanceRepository interface {
	SaveAdvance(ctx context.Context, advance domain.Advance) (string, error)
	UpdateAdvance(ctx context.Context, advance domain.Advance) error
	FindAdvanceById(ctx context.Context, advanceId string) (domain.Advance, error)
	FindAllAdvances(ctx context.Context, filterOptions domain.AdvancesFilterOptions) ([]domain.Advance, int, error)
	GetAdvanceSumByStatus(ctx context.Context, userID string, status domain.RequestStatus) (float64, error)
	GetReconciledAdvancesSum(ctx context.Context, userID string) (float64, error)
}
