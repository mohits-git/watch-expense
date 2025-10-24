package authctx

import (
	"context"
	"testing"
)

func Test_authctx_WithToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add token to context",
			args: args{
				ctx:   context.Background(),
				token: "sample_token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithToken(tt.args.ctx, tt.args.token)
			gotToken, ok := TokenFromCtx(ctx)
			if !ok {
				t.Errorf("TokenFromCtx() returned !ok")
			}
			if gotToken != tt.args.token {
				t.Errorf("TokenFromCtx() = %v, want %v", gotToken, tt.args.token)
			}
		})
	}
}

func Test_authctx_TokenFromCtx(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		args   args
		want   string
		wantOk bool
	}{
		{
			name: "token present in context",
			args: args{
				ctx: WithToken(context.Background(), "sample_token"),
			},
			want:   "sample_token",
			wantOk: true,
		},
		{
			name: "no token in context",
			args: args{
				ctx: context.Background(),
			},
			want:   "",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := TokenFromCtx(tt.args.ctx)
			if got != tt.want {
				t.Errorf("TokenFromCtx() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("TokenFromCtx() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
