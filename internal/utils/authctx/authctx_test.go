package authctx

import (
	"context"
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
)

func Test_authctx_NewUserClaims(t *testing.T) {
	type args struct {
		userID string
		role   domain.UserRole
	}
	tests := []struct {
		name string
		args args
		want UserClaims
	}{
		{
			name: "create user claims",
			args: args{
				userID: "1",
				role:   domain.Employee,
			},
			want: UserClaims{
				UserID: "1",
				Role:   domain.Employee,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUserClaims(tt.args.userID, tt.args.role); got != tt.want {
				t.Errorf("NewUserClaims() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authctx_WithUserClaims(t *testing.T) {
	type args struct {
		ctx    context.Context
		claims *UserClaims
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add user claims to context",
			args: args{
				ctx:    context.Background(),
				claims: &UserClaims{UserID: "1", Role: domain.Employee},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserClaims(tt.args.ctx, tt.args.claims)
			if ctx == nil {
				t.Errorf("WithUserClaims() returned nil context")
			}
		})
	}
}

func Test_authctx_UserClaimsFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *UserClaims
		wantOk  bool
		setupFn func() context.Context
	}{
		{
			name: "context with user claims",
			args: args{
				ctx: nil,
			},
			want:   &UserClaims{UserID: "1", Role: domain.Employee},
			wantOk: true,
			setupFn: func() context.Context {
				return WithUserClaims(context.Background(), &UserClaims{UserID: "1", Role: domain.Employee})
			},
		},
		{
			name: "context without user claims",
			args: args{
				ctx: context.Background(),
			},
			want:   nil,
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.args.ctx = tt.setupFn()
			}
			got, gotOk := UserClaimsFromCtx(tt.args.ctx)
			if gotOk != tt.wantOk {
				t.Errorf("UserClaimsFromCtx() gotOk = %v, want %v", gotOk, tt.wantOk)
				return
			}
			if (got == nil) != (tt.want == nil) || (got != nil && *got != *tt.want) {
				t.Errorf("UserClaimsFromCtx() got = %v, want %v", got, tt.want)
			}
		})
	}
}
