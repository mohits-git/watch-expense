package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validator_ValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "valid - standard email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "valid - email with subdomain",
			email: "user@mail.example.com",
			want:  true,
		},
		{
			name:  "valid - email with plus sign",
			email: "user+test@example.com",
			want:  true,
		},
		{
			name:  "valid - email with dots",
			email: "first.last@example.com",
			want:  true,
		},
		{
			name:  "valid - email with numbers",
			email: "user123@example123.com",
			want:  true,
		},
		{
			name:  "valid - email with underscore",
			email: "user_name@example.com",
			want:  true,
		},
		{
			name:  "valid - email with hyphen in domain",
			email: "user@ex-ample.com",
			want:  true,
		},
		{
			name:  "invalid - missing @ symbol",
			email: "userexample.com",
			want:  false,
		},
		{
			name:  "invalid - missing domain",
			email: "user@",
			want:  false,
		},
		{
			name:  "invalid - missing local part",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "invalid - missing TLD",
			email: "user@example",
			want:  false,
		},
		{
			name:  "invalid - double @ symbol",
			email: "user@@example.com",
			want:  false,
		},
		{
			name:  "invalid - spaces",
			email: "user @example.com",
			want:  false,
		},
		{
			name:  "invalid - special characters",
			email: "user!#$@example.com",
			want:  false,
		},
		{
			name:  "invalid - empty string",
			email: "",
			want:  false,
		},
		{
			name:  "invalid - too short (less than 3 chars)",
			email: "a@",
			want:  false,
		},
		{
			name:  "invalid - too long (more than 254 chars)",
			email: "verylongemailaddress" + string(make([]byte, 250)) + "@example.com",
			want:  false,
		},
		{
			name:  "invalid - single letter TLD",
			email: "user@example.c",
			want:  false,
		},
		{
			name:  "valid - minimum length",
			email: "a@b.co",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			assert.Equal(t, tt.want, result)
		})
	}
}
