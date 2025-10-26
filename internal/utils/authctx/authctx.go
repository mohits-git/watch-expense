package authctx

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type userKeyType string

const userKey userKeyType = "user"

type UserClaims struct {
	UserID string
	Role   domain.UserRole
	Name   string
	Email  string
}

func NewUserClaims(userID, name, email string, role domain.UserRole) UserClaims {
	return UserClaims{
		UserID: userID,
		Role:   role,
		Name:   name,
		Email:  email,
	}
}

func WithUserClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, userKey, claims)
}

func UserClaimsFromCtx(ctx context.Context) (*UserClaims, bool) {
	val := ctx.Value(userKey)
	if val == nil {
		return nil, false
	}

	claims, ok := val.(*UserClaims)
	return claims, ok
}
