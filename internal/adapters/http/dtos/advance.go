package dtos

import "github.com/mohits-git/watch-expense/internal/domain"

type Advance struct {
	ID                   string
	UserID               string
	Amount               float64
	Purpose              string
	Description          string
	Status               domain.RequestStatus
	ReconciledExpenseID string
	ApprovedBy           string
	ApprovedAt           int64
	ReviewedBy           string
	ReviewedAt           int64
	CreatedAt            int64
	UpdatedAt            int64
}
