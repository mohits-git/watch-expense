package domain

type Advance struct {
	ID                   string        `json:"id"`
	UserID               string        `json:"userId"`
	Amount               float64       `json:"amount"`
  Purpose              string        `json:"purpose"`
	Description          string        `json:"description"`
	Status               RequestStatus `json:"status"`
	ReconcilledExpenseID string        `json:"reconcilledExpenseId"`
	ApprovedBy           string        `json:"approvedBy"`
	ApprovedAt           int64         `json:"approvedAt"`
	ReviewedBy           string        `json:"reviewedBy"`
	ReviewedAt           int64         `json:"reviewedAt"`
	CreatedAt            int64         `json:"createdAt"`
	UpdatedAt            int64         `json:"updatedAt"`
}
