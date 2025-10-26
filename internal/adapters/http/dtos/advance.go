package dtos

import "github.com/mohits-git/watch-expense/internal/domain"

type Advance struct {
	ID                  string               `json:"id"`
	UserID              string               `json:"userId"`
	Amount              float64              `json:"amount"`
	Purpose             string               `json:"purpose"`
	Description         string               `json:"description"`
	Status              domain.RequestStatus `json:"status"`
	ReconciledExpenseID string               `json:"reconciledExpenseId"`
	ApprovedBy          string               `json:"approvedBy"`
	ApprovedAt          int64                `json:"approvedAt"`
	ReviewedBy          string               `json:"reviewedBy"`
	ReviewedAt          int64                `json:"reviewedAt"`
	CreatedAt           int64                `json:"createdAt"`
	UpdatedAt           int64                `json:"updatedAt"`
}

type CreateAdvanceRequest struct {
	Amount      float64 `json:"amount"`
	Purpose     string  `json:"purpose"`
	Description string  `json:"description"`
}

type CreateAdvanceResponse struct {
	ID string `json:"id"`
}

type UpdateAdvanceRequest struct {
	Amount      float64 `json:"amount"`
	Purpose     string  `json:"purpose"`
	Description string  `json:"description"`
}

type UpdateAdvanceStatusRequest struct {
	Status domain.RequestStatus `json:"status"`
}

type GetAdvancesResponse struct {
	TotalAdvances int       `json:"totalAdvances"`
	Advances      []Advance `json:"advances"`
}

func ToAdvanceDTO(a domain.Advance) Advance {
	return Advance{
		ID:                  a.ID,
		UserID:              a.UserID,
		Amount:              a.Amount,
		Purpose:             a.Purpose,
		Description:         a.Description,
		Status:              a.Status,
		ReconciledExpenseID: a.ReconciledExpenseID,
		ApprovedBy:          a.ApprovedBy,
		ApprovedAt:          a.ApprovedAt,
		ReviewedBy:          a.ReviewedBy,
		ReviewedAt:          a.ReviewedAt,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}

func ToAdvanceDomain(dto Advance) domain.Advance {
	return domain.Advance{
		ID:                  dto.ID,
		UserID:              dto.UserID,
		Amount:              dto.Amount,
		Purpose:             dto.Purpose,
		Description:         dto.Description,
		Status:              dto.Status,
		ReconciledExpenseID: dto.ReconciledExpenseID,
		ApprovedBy:          dto.ApprovedBy,
		ApprovedAt:          dto.ApprovedAt,
		ReviewedBy:          dto.ReviewedBy,
		ReviewedAt:          dto.ReviewedAt,
		CreatedAt:           dto.CreatedAt,
		UpdatedAt:           dto.UpdatedAt,
	}
}

func ToAdvancesDTOs(domains []domain.Advance) []Advance {
	advances := make([]Advance, len(domains))
	for i, domainAdvance := range domains {
		advances[i] = ToAdvanceDTO(domainAdvance)
	}
	return advances
}

type AdvanceSummary struct {
	Approved   float64 `json:"approved"`
	Reconciled float64 `json:"reconciled"`
	Pending    float64 `json:"pendingReconciliation"`
	Rejected   float64 `json:"rejectedAdvance"`
}
