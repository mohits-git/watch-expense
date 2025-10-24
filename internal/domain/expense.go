package domain

type RequestStatus string

const (
	Pending  RequestStatus = "PENDING"
	Approved RequestStatus = "APPROVED"
	Rejected RequestStatus = "REJECTED"
	Reviewed RequestStatus = "REVIEWED"
)

type Expense struct {
	ID            string
	UserID        string
	Amount        float64
	Description   string
	Status        RequestStatus
	Purpose       string
	ApprovedBy    string
	ApprovedAt    int64
	ReviewedBy    string
	ReviewedAt    int64
	IsReconcilled bool
	Bills         []Bill
	CreatedAt     int64
	UpdatedAt     int64
}

type Bill struct {
	ID            string
	ExpenseID     string
	Amount        float64
	Description   string
	AttachmentURL string
}
