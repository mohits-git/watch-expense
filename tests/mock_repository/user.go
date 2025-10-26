package mockrepository

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/stretchr/testify/mock"
)

type UserRepository struct {
	mock.Mock
}

func NewMockUserRepository() ports.UserRepository {
	return &UserRepository{}
}

func (u *UserRepository) SaveUser(ctx context.Context, user domain.User) (string, error) {
	args := u.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (u *UserRepository) UpdateUser(ctx context.Context, user domain.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserRepository) FindUserById(ctx context.Context, userId string) (domain.User, error) {
	args := u.Called(ctx, userId)
	return args.Get(0).(domain.User), args.Error(1)
}

func (u *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	args := u.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (u *UserRepository) FindAllUsers(ctx context.Context) ([]domain.User, error) {
	args := u.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (u *UserRepository) DeleteUser(ctx context.Context, userID string) error {
  args := u.Called(ctx, userID)
  return args.Error(0)
}
