package dtos

import "github.com/mohits-git/watch-expense/internal/domain"

type Expense struct {
	ID           string               `json:"id"`
	UserID       string               `json:"user_id"`
	Amount       float64              `json:"amount"`
	Description  string               `json:"description"`
	Status       domain.RequestStatus `json:"status"`
	Purpose      string               `json:"purpose"`
	ApprovedBy   string               `json:"approvedBy"`
	ApprovedAt   int64                `json:"approvedAt"`
	ReviewedBy   string               `json:"reviewedBy"`
	ReviewedAt   int64                `json:"reviewedAt"`
	IsReconciled bool                 `json:"isReconciled"`
	Bills        []Bill               `json:"bills"`
	CreatedAt    int64                `json:"createdAt"`
	UpdatedAt    int64                `json:"updatedAt"`
}

type Bill struct {
	ID            string  `json:"id"`
	ExpenseID     string  `json:"expenseId"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	AttachmentURL string  `json:"attachmentUrl"`
}
