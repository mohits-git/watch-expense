package domain

type Advance struct {
	ID                   string
	UserID               string
	Amount               float64
	Purpose              string
	Description          string
	Status               RequestStatus
	ReconciledExpenseID string
	ApprovedBy           string
	ApprovedAt           int64
	ReviewedBy           string
	ReviewedAt           int64
	CreatedAt            int64
	UpdatedAt            int64
}

type AdvancesFilterOptions struct {
	UserID string
	Status RequestStatus
	Page   int
	Limit  int
}
