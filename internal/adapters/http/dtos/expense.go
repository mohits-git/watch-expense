package dtos

import (
	"github.com/mohits-git/watch-expense/internal/domain"
)

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
	AdvanceID    string               `json:"advanceId"`
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

type CreateBillRequest struct {
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	AttachmentURL string  `json:"attachmentUrl"`
}

type CreateExpenseRequest struct {
	Amount       float64             `json:"amount"`
	Description  string              `json:"description"`
	Purpose      string              `json:"purpose"`
	IsReconciled bool                `json:"isReconciled"`
	AdvanceID    string              `json:"advanceId"`
	Bills        []CreateBillRequest `json:"bills"`
}

type CreateExpenseResponse struct {
	ID string `json:"id"`
}

type GetExpensesResponse struct {
	TotalExpenses int       `json:"totalExpenses"`
	Expenses      []Expense `json:"expenses"`
}

type UpdateExpenseRequest struct {
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Purpose      string  `json:"purpose"`
	IsReconciled bool    `json:"isReconciled"`
}

type UpdateExpenseStatusRequest struct {
	Status domain.RequestStatus `json:"status"`
}

func ToExpenseDomain(dto Expense) domain.Expense {
	bills := make([]domain.Bill, len(dto.Bills))
	for i, billDTO := range dto.Bills {
		bills[i] = domain.Bill{
			ID:            billDTO.ID,
			ExpenseID:     billDTO.ExpenseID,
			Amount:        billDTO.Amount,
			Description:   billDTO.Description,
			AttachmentURL: billDTO.AttachmentURL,
		}
	}

	return domain.Expense{
		ID:           dto.ID,
		UserID:       dto.UserID,
		Amount:       dto.Amount,
		Description:  dto.Description,
		Status:       dto.Status,
		Purpose:      dto.Purpose,
		ApprovedBy:   dto.ApprovedBy,
		ApprovedAt:   dto.ApprovedAt,
		ReviewedBy:   dto.ReviewedBy,
		ReviewedAt:   dto.ReviewedAt,
		IsReconciled: dto.IsReconciled,
		AdvanceID:    dto.AdvanceID,
		Bills:        bills,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}
}

func ToExpenseDTO(domainExpense domain.Expense) Expense {
	bills := make([]Bill, len(domainExpense.Bills))
	for i, billDomain := range domainExpense.Bills {
		bills[i] = Bill{
			ID:            billDomain.ID,
			ExpenseID:     billDomain.ExpenseID,
			Amount:        billDomain.Amount,
			Description:   billDomain.Description,
			AttachmentURL: billDomain.AttachmentURL,
		}
	}

	return Expense{
		ID:           domainExpense.ID,
		UserID:       domainExpense.UserID,
		Amount:       domainExpense.Amount,
		Description:  domainExpense.Description,
		Status:       domainExpense.Status,
		Purpose:      domainExpense.Purpose,
		ApprovedBy:   domainExpense.ApprovedBy,
		ApprovedAt:   domainExpense.ApprovedAt,
		ReviewedBy:   domainExpense.ReviewedBy,
		ReviewedAt:   domainExpense.ReviewedAt,
		IsReconciled: domainExpense.IsReconciled,
		AdvanceID:    domainExpense.AdvanceID,
		Bills:        bills,
		CreatedAt:    domainExpense.CreatedAt,
		UpdatedAt:    domainExpense.UpdatedAt,
	}
}

func ToCreateExpenseDomain(dto CreateExpenseRequest) domain.Expense {
	bills := make([]domain.Bill, len(dto.Bills))
	for i, billDTO := range dto.Bills {
		bills[i] = domain.Bill{
			Amount:        billDTO.Amount,
			Description:   billDTO.Description,
			AttachmentURL: billDTO.AttachmentURL,
		}
	}

	return domain.Expense{
		Amount:       dto.Amount,
		Description:  dto.Description,
		Purpose:      dto.Purpose,
		IsReconciled: dto.IsReconciled,
		AdvanceID:    dto.AdvanceID,
		Bills:        bills,
	}
}

type ExpenseSummary struct {
	TotalExpenses     int `json:"totalExpense"`
	PendingExpense    int `json:"pendingExpense"`
	ReimbursedExpense int `json:"reimbursedExpense"`
	RejectedExpense   int `json:"rejectedExpense"`
}
