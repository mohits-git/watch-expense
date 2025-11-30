package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateDepartmentCreation(t *testing.T) {
	tests := []struct {
		name       string
		department domain.Department
		want       bool
	}{
		{
			name: "valid - department with positive budget",
			department: domain.Department{
				Name:   "Engineering",
				Budget: 100000.00,
			},
			want: true,
		},
		{
			name: "valid - department with zero budget",
			department: domain.Department{
				Name:   "New Department",
				Budget: 0,
			},
			want: true,
		},
		{
			name: "valid - department with decimal budget",
			department: domain.Department{
				Name:   "Marketing",
				Budget: 50000.50,
			},
			want: true,
		},
		{
			name: "invalid - empty name",
			department: domain.Department{
				Name:   "",
				Budget: 10000,
			},
			want: false,
		},
		{
			name: "invalid - negative budget",
			department: domain.Department{
				Name:   "Finance",
				Budget: -1000,
			},
			want: false,
		},
		{
			name: "invalid - empty name and negative budget",
			department: domain.Department{
				Name:   "",
				Budget: -500,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDepartmentCreation(tt.department)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateDepartmentUpdate(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name       string
		department domain.Department
		want       bool
	}{
		{
			name: "valid - update department",
			department: domain.Department{
				ID:     validUUID,
				Name:   "Engineering Updated",
				Budget: 150000.00,
			},
			want: true,
		},
		{
			name: "valid - update with zero budget",
			department: domain.Department{
				ID:     validUUID,
				Name:   "Department",
				Budget: 0,
			},
			want: true,
		},
		{
			name: "invalid - invalid UUID",
			department: domain.Department{
				ID:     "invalid-uuid",
				Name:   "Engineering",
				Budget: 100000,
			},
			want: false,
		},
		{
			name: "invalid - empty UUID",
			department: domain.Department{
				ID:     "",
				Name:   "Engineering",
				Budget: 100000,
			},
			want: false,
		},
		{
			name: "invalid - empty name",
			department: domain.Department{
				ID:     validUUID,
				Name:   "",
				Budget: 100000,
			},
			want: false,
		},
		{
			name: "invalid - negative budget",
			department: domain.Department{
				ID:     validUUID,
				Name:   "Engineering",
				Budget: -5000,
			},
			want: false,
		},
		{
			name: "invalid - all invalid fields",
			department: domain.Department{
				ID:     "invalid",
				Name:   "",
				Budget: -1000,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDepartmentUpdate(tt.department)
			assert.Equal(t, tt.want, result)
		})
	}
}
