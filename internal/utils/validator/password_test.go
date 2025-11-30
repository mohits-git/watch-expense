package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "valid - exactly 8 characters",
			password: "password",
			want:     true,
		},
		{
			name:     "valid - more than 8 characters",
			password: "password123",
			want:     true,
		},
		{
			name:     "valid - long password",
			password: "thisIsAVeryLongPasswordWithLotsOfCharacters",
			want:     true,
		},
		{
			name:     "valid - password with special characters",
			password: "P@ssw0rd!",
			want:     true,
		},
		{
			name:     "valid - password with spaces",
			password: "my password",
			want:     true,
		},
		{
			name:     "invalid - 7 characters",
			password: "passwor",
			want:     false,
		},
		{
			name:     "invalid - 6 characters",
			password: "passwd",
			want:     false,
		},
		{
			name:     "invalid - empty string",
			password: "",
			want:     false,
		},
		{
			name:     "invalid - 1 character",
			password: "p",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.password)
			assert.Equal(t, tt.want, result)
		})
	}
}
