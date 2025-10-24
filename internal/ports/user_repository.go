package ports

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
)

type UserRepository interface {
  SaveUser(ctx context.Context, user domain.User) (string, error)
  UpdateUser(ctx context.Context, user domain.User) error
  FindUserById(ctx context.Context, userId string) (domain.User, error)
  FindUserByEmail(ctx context.Context, email string) (domain.User, error)
  FindAllUsers(ctx context.Context) ([]domain.User, error)
}
