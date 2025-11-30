package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateProjectCreation(t *testing.T) {
	validDeptID := uuid.New().String()

	tests := []struct {
		name    string
		project domain.Project
		want    bool
	}{
		{
			name: "valid - project without department",
			project: domain.Project{
				Name:   "New Website",
				Budget: 50000.00,
			},
			want: true,
		},
		{
			name: "valid - project with department",
			project: domain.Project{
				Name:         "Mobile App",
				Budget:       75000.00,
				DepartmentID: validDeptID,
			},
			want: true,
		},
		{
			name: "valid - project with zero budget",
			project: domain.Project{
				Name:   "Research Project",
				Budget: 0,
			},
			want: true,
		},
		{
			name: "valid - project with decimal budget",
			project: domain.Project{
				Name:   "Marketing Campaign",
				Budget: 25000.50,
			},
			want: true,
		},
		{
			name: "invalid - empty name",
			project: domain.Project{
				Name:   "",
				Budget: 10000,
			},
			want: false,
		},
		{
			name: "invalid - negative budget",
			project: domain.Project{
				Name:   "Project",
				Budget: -1000,
			},
			want: false,
		},
		{
			name: "invalid - invalid department ID",
			project: domain.Project{
				Name:         "Project",
				Budget:       10000,
				DepartmentID: "invalid-dept-id",
			},
			want: false,
		},
		{
			name: "invalid - empty name and negative budget",
			project: domain.Project{
				Name:   "",
				Budget: -500,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateProjectCreation(tt.project)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateProjectUpdate(t *testing.T) {
	validUUID := uuid.New().String()
	validDeptID := uuid.New().String()

	tests := []struct {
		name    string
		project domain.Project
		want    bool
	}{
		{
			name: "valid - update project without department",
			project: domain.Project{
				ID:     validUUID,
				Name:   "Updated Project",
				Budget: 60000.00,
			},
			want: true,
		},
		{
			name: "valid - update project with department",
			project: domain.Project{
				ID:           validUUID,
				Name:         "Updated Project",
				Budget:       80000.00,
				DepartmentID: validDeptID,
			},
			want: true,
		},
		{
			name: "valid - update with zero budget",
			project: domain.Project{
				ID:     validUUID,
				Name:   "Project",
				Budget: 0,
			},
			want: true,
		},
		{
			name: "invalid - invalid UUID",
			project: domain.Project{
				ID:     "invalid-uuid",
				Name:   "Project",
				Budget: 10000,
			},
			want: false,
		},
		{
			name: "invalid - empty UUID",
			project: domain.Project{
				ID:     "",
				Name:   "Project",
				Budget: 10000,
			},
			want: false,
		},
		{
			name: "invalid - empty name",
			project: domain.Project{
				ID:     validUUID,
				Name:   "",
				Budget: 10000,
			},
			want: false,
		},
		{
			name: "invalid - negative budget",
			project: domain.Project{
				ID:     validUUID,
				Name:   "Project",
				Budget: -5000,
			},
			want: false,
		},
		{
			name: "invalid - invalid department ID",
			project: domain.Project{
				ID:           validUUID,
				Name:         "Project",
				Budget:       10000,
				DepartmentID: "invalid-dept",
			},
			want: false,
		},
		{
			name: "invalid - all invalid fields",
			project: domain.Project{
				ID:           "invalid",
				Name:         "",
				Budget:       -1000,
				DepartmentID: "invalid",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateProjectUpdate(tt.project)
			assert.Equal(t, tt.want, result)
		})
	}
}
