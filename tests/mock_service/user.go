package mockservice

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/mock"
)

type UserService struct {
	mock.Mock
}

func NewMockUserService() *UserService {
	return &UserService{}
}

func (u *UserService) CreateUser(ctx context.Context, user domain.User) (string, error) {
	args := u.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (u *UserService) UpdateUser(ctx context.Context, user domain.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserService) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	args := u.Called(ctx, userID)
	return args.Get(0).(domain.User), args.Error(1)
}

func (u *UserService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	args := u.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (u *UserService) DeleteUser(ctx context.Context, userID string) error {
	args := u.Called(ctx, userID)
	return args.Error(0)
}

func (u *UserService) GetUserBudget(ctx context.Context) (float64, error) {
	args := u.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}
