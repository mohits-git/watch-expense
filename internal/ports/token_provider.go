package ports

import "github.com/mohits-git/watch-expense/internal/utils/authctx"

type TokenProvider interface {
  GenerateToken(claims authctx.UserClaims) (token string, err error)
  ValidateToken(token string) (claims authctx.UserClaims, err error)
}
