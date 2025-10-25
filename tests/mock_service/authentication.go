package mockservice

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/stretchr/testify/mock"
)

type AuthenticationService struct {
	mock.Mock
}

func NewMockAuthenticationService() *AuthenticationService {
	return &AuthenticationService{}
}

func (a *AuthenticationService) Login(ctx context.Context, email, password string) (string, error) {
	args := a.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (a *AuthenticationService) Logout(ctx context.Context, token string) error {
	args := a.Called(ctx, token)
	return args.Error(0)
}

func (a *AuthenticationService) GetCurrentUser(ctx context.Context) (domain.User, error) {
	args := a.Called(ctx)
	return args.Get(0).(domain.User), args.Error(1)
}
