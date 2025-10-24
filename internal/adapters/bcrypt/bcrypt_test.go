package bcrypt

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func Test_bcrypt_NewBcryptPasswordHasher(t *testing.T) {
	tests := []struct {
		name string
		cost int
		want int
	}{
		{
			name: "valid cost",
			cost: 10,
			want: bcrypt.DefaultCost,
		},
		{
			name: "cost less than min cost",
			cost: 1,
			want: bcrypt.MinCost,
		},
		{
			name: "cost greater than max cost",
			cost: 20000,
			want: bcrypt.MaxCost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBcryptPasswordHasher(tt.cost)
			if hasher.cost != tt.want {
				t.Errorf("NewBcryptPasswordHasher() cost = %v, want %v", hasher.cost, tt.want)
			}
		})
	}
}

func Test_bcrypt_HashPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher(bcrypt.DefaultCost)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mysecretpassword",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "very long password",
			password: string(make([]byte, 1000)),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hasher.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("HashPassword() got empty hash")
			}
		})
	}
}

func Test_bcrypt_ComparePassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher(bcrypt.DefaultCost)
	password := "mysecretpassword"
	validHash := "$2a$10$o91wsoiH0Mx.9ESJOmZtj.OxMrLeXELxsvsuaEl.Tc9bgjSsE3bD."

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		want           bool
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: validHash,
			password:       password,
			want:           true,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: validHash,
			password:       "wrongpassword",
			want:           false,
			wantErr:        false,
		},
		{
			name:           "empty hashed password",
			hashedPassword: "",
			password:       "",
			want:           false,
			wantErr:        true,
		},
		{
			name:           "too long password",
			hashedPassword: validHash,
			password:       string(make([]byte, 1000)),
			want:           false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hasher.ComparePassword(tt.hashedPassword, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ComparePassword() got = %v, want %v", got, tt.want)
			}
		})
	}
}
