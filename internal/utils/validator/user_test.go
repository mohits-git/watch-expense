package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateUserCreation(t *testing.T) {
	tests := []struct {
		name string
		user domain.User
		want bool
	}{
		{
			name: "valid - employee user",
			user: domain.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     domain.Employee,
			},
			want: true,
		},
		{
			name: "valid - manager user",
			user: domain.User{
				Name:     "Jane Smith",
				Email:    "jane@example.com",
				Password: "managerpass",
				Role:     domain.Manager,
			},
			want: true,
		},
		{
			name: "valid - admin user",
			user: domain.User{
				Name:     "Admin User",
				Email:    "admin@example.com",
				Password: "adminpass123",
				Role:     domain.Admin,
			},
			want: true,
		},
		{
			name: "invalid - empty name",
			user: domain.User{
				Name:     "",
				Email:    "user@example.com",
				Password: "password123",
				Role:     domain.Employee,
			},
			want: false,
		},
		{
			name: "invalid - invalid email",
			user: domain.User{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
				Role:     domain.Employee,
			},
			want: false,
		},
		{
			name: "invalid - invalid role",
			user: domain.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     "InvalidRole",
			},
			want: false,
		},
		{
			name: "invalid - empty role",
			user: domain.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUserCreation(tt.user)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateUserUpdate(t *testing.T) {
	validUUID := uuid.New().String()
	validDeptID := uuid.New().String()
	validProjID := uuid.New().String()

	tests := []struct {
		name string
		user domain.User
		want bool
	}{
		{
			name: "valid - basic update",
			user: domain.User{
				ID:       validUUID,
				Name:     "John Doe Updated",
				Email:    "john.updated@example.com",
				Password: "newpassword123",
				Role:     domain.Manager,
			},
			want: true,
		},
		{
			name: "valid - update with department",
			user: domain.User{
				ID:           validUUID,
				Name:         "John Doe",
				Email:        "john@example.com",
				Password:     "password123",
				Role:         domain.Employee,
				DepartmentID: validDeptID,
			},
			want: true,
		},
		{
			name: "valid - update with project",
			user: domain.User{
				ID:        validUUID,
				Name:      "John Doe",
				Email:     "john@example.com",
				Password:  "password123",
				Role:      domain.Employee,
				ProjectID: validProjID,
			},
			want: true,
		},
		{
			name: "valid - update with both department and project",
			user: domain.User{
				ID:           validUUID,
				Name:         "John Doe",
				Email:        "john@example.com",
				Password:     "password123",
				Role:         domain.Employee,
				DepartmentID: validDeptID,
				ProjectID:    validProjID,
			},
			want: true,
		},
		{
			name: "invalid - invalid UUID",
			user: domain.User{
				ID:       "invalid-uuid",
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     domain.Employee,
			},
			want: false,
		},
		{
			name: "invalid - empty name",
			user: domain.User{
				ID:       validUUID,
				Name:     "",
				Email:    "john@example.com",
				Password: "password123",
				Role:     domain.Employee,
			},
			want: false,
		},
		{
			name: "invalid - invalid email",
			user: domain.User{
				ID:       validUUID,
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
				Role:     domain.Employee,
			},
			want: false,
		},
		{
			name: "invalid - short password",
			user: domain.User{
				ID:       validUUID,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "short",
				Role:     domain.Employee,
			},
			want: false,
		},
		{
			name: "invalid - invalid role",
			user: domain.User{
				ID:       validUUID,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     "InvalidRole",
			},
			want: false,
		},
		{
			name: "invalid - invalid department ID",
			user: domain.User{
				ID:           validUUID,
				Name:         "John Doe",
				Email:        "john@example.com",
				Password:     "password123",
				Role:         domain.Employee,
				DepartmentID: "invalid-dept-id",
			},
			want: false,
		},
		{
			name: "invalid - invalid project ID",
			user: domain.User{
				ID:        validUUID,
				Name:      "John Doe",
				Email:     "john@example.com",
				Password:  "password123",
				Role:      domain.Employee,
				ProjectID: "invalid-proj-id",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUserUpdate(tt.user)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_validator_ValidateUserRole(t *testing.T) {
	tests := []struct {
		name string
		role domain.UserRole
		want bool
	}{
		{
			name: "valid - employee role",
			role: domain.Employee,
			want: true,
		},
		{
			name: "valid - manager role",
			role: domain.Manager,
			want: true,
		},
		{
			name: "valid - admin role",
			role: domain.Admin,
			want: true,
		},
		{
			name: "invalid - empty role",
			role: "",
			want: false,
		},
		{
			name: "invalid - unknown role",
			role: "superuser",
			want: false,
		},
		{
			name: "invalid - case mismatch",
			role: "employee",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUserRole(tt.role)
			assert.Equal(t, tt.want, result)
		})
	}
}
