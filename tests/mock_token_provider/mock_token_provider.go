package mocktokenprovider

import (
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	"github.com/stretchr/testify/mock"
)

type TokenProvider struct {
	mock.Mock
}

func (t *TokenProvider) GenerateToken(claims authctx.UserClaims) (token string, err error) {
	args := t.Called(claims)
	return args.String(0), args.Error(1)
}

func (t *TokenProvider) ValidateToken(token string) (claims authctx.UserClaims, err error) {
	args := t.Called(token)
	return args.Get(0).(authctx.UserClaims), args.Error(1)
}
