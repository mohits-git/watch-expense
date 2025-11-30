package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateAdvanceCreation(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name    string
		advance domain.Advance
		want    bool
	}{
		{
			name: "valid - advance with minimum fields",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "Business Trip",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - advance with decimal amount",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  500.50,
				Purpose: "Office Supplies",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - advance with large amount",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  50000.00,
				Purpose: "Equipment Purchase",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "invalid - zero amount",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  0,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - negative amount",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  -500.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty user ID",
			advance: domain.Advance{
				UserID:  "",
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - invalid user ID",
			advance: domain.Advance{
				UserID:  "invalid-user-id",
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty purpose",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - status not pending",
			advance: domain.Advance{
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Approved,
			},
			want: false,
		},
		{
			name: "invalid - multiple validation failures",
			advance: domain.Advance{
				UserID:  "",
				Amount:  -1000,
				Purpose: "",
				Status:  domain.Rejected,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAdvanceCreation(tt.advance)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateAdvanceUpdate(t *testing.T) {
	validUUID := uuid.New().String()
	validUserID := uuid.New().String()

	tests := []struct {
		name    string
		advance domain.Advance
		want    bool
	}{
		{
			name: "valid - update advance with all fields",
			advance: domain.Advance{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  2000.00,
				Purpose: "Updated Trip",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - update advance without user ID",
			advance: domain.Advance{
				ID:      validUUID,
				Amount:  1500.00,
				Purpose: "Updated Purpose",
				Status:  domain.Pending,
			},
			want: true,
		},
		{
			name: "valid - update advance with empty status",
			advance: domain.Advance{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  "",
			},
			want: true,
		},
		{
			name: "invalid - invalid advance ID",
			advance: domain.Advance{
				ID:      "invalid-id",
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty advance ID",
			advance: domain.Advance{
				ID:      "",
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - zero amount",
			advance: domain.Advance{
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
			advance: domain.Advance{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  -1000,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - invalid user ID when provided",
			advance: domain.Advance{
				ID:      validUUID,
				UserID:  "invalid-user",
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - empty purpose",
			advance: domain.Advance{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "",
				Status:  domain.Pending,
			},
			want: false,
		},
		{
			name: "invalid - status not pending when provided",
			advance: domain.Advance{
				ID:      validUUID,
				UserID:  validUserID,
				Amount:  1000.00,
				Purpose: "Travel",
				Status:  domain.Approved,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAdvanceUpdate(tt.advance)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateAdvanceStatus(t *testing.T) {
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
			status: "in-progress",
			want:   false,
		},
		{
			name:   "invalid - case mismatch",
			status: "approved",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAdvanceStatus(tt.status)
			assert.Equal(t, tt.want, result)
		})
	}
}
