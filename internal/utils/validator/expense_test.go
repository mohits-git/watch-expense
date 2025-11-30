package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateExpenseCreation(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name    string
		expense domain.Expense
		want    bool
	}{
		{
			name: "valid - expense with minimum fields",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - expense with decimal amount",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  99.99,
				Purpose: "Office Supplies",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - expense with large amount",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  10000.50,
				Purpose: "Equipment",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "invalid - zero amount",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  0,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - negative amount",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  -50.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty user ID",
			expense: domain.Expense{
				UserID:  "",
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - invalid user ID",
			expense: domain.Expense{
				UserID:  "invalid-user-id",
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty purpose",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - status not pending",
			expense: domain.Expense{
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Approved,
			},
			want: false,
		},
		{
			name: "invalid - multiple validation failures",
			expense: domain.Expense{
				UserID:  "",
				Amount:  -100,
				Purpose: "",
				Status:  domain.Rejected,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateExpenseCreation(tt.expense)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateExpenseUpdate(t *testing.T) {
	validUUID := uuid.New().String()
	validUserID := uuid.New().String()

	tests := []struct {
		name    string
		expense domain.Expense
		want    bool
	}{
		{
			name: "valid - update expense",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  150.00,
				Purpose: "Updated Travel",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - update with decimal amount",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  123.45,
				Purpose: "Updated Supplies",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "invalid - invalid expense ID",
			expense: domain.Expense{
				ID:      "invalid-id",
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty expense ID",
			expense: domain.Expense{
				ID:      "",
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - zero amount",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  0,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - negative amount",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  -100,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty user ID",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  "",
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - invalid user ID",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  "invalid-user",
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty purpose",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - status not pending",
			expense: domain.Expense{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  100.00,
				Purpose: "Travel",
				Status:  domain.Approved,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateExpenseUpdate(tt.expense)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateExpenseStatus(t *testing.T) {
	tests := []struct {
		name   string
		status domain.RequestStatus
		want   bool
	}{
		{
			name:   "valid - pending status",
			status: domain.Pending,
			want:   true,
		},
		{
			name:   "valid - approved status",
			status: domain.Approved,
			want:   true,
		},
		{
			name:   "valid - rejected status",
			status: domain.Rejected,
			want:   true,
		},
		{
			name:   "valid - reviewed status",
			status: domain.Reviewed,
			want:   true,
		},
		{
			name:   "invalid - empty status",
			status: "",
			want:   false,
		},
		{
			name:   "invalid - unknown status",
			status: "processing",
			want:   false,
		},
		{
			name:   "invalid - case mismatch",
			status: "pending",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateExpenseStatus(tt.status)
			assert.Equal(t, tt.want, result)
		})
	}
}
