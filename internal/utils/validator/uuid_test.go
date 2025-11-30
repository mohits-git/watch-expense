package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateUUID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "valid - UUID v4",
			id:   uuid.New().String(),
			want: true,
		},
		{
			name: "valid - lowercase UUID",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			want: true,
		},
		{
			name: "valid - uppercase UUID",
			id:   "550E8400-E29B-41D4-A716-446655440000",
			want: true,
		},
		{
			name: "valid - mixed case UUID",
			id:   "550e8400-E29B-41d4-A716-446655440000",
			want: true,
		},
		{
			name: "invalid - wrong format",
			id:   "550e8400-e29b-41d4-a716-44665544000",
			want: false,
		},
		{
			name: "invalid - too short",
			id:   "550e8400-e29b-41d4",
			want: false,
		},
		{
			name: "invalid - too long",
			id:   "550e8400-e29b-41d4-a716-446655440000-extra",
			want: false,
		},
		{
			name: "invalid - contains invalid characters",
			id:   "550e8400-e29b-41d4-a716-44665544000g",
			want: false,
		},
		{
			name: "invalid - empty string",
			id:   "",
			want: false,
		},
		{
			name: "invalid - random string",
			id:   "not-a-uuid",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUUID(tt.id)
			assert.Equal(t, tt.want, result)
		})
	}
}
