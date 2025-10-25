package services

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type UserService interface {
  CreateUser(ctx context.Context, user domain.User) (string, error)
  GetUserByID(ctx context.Context, userID string) (domain.User, error)
  GetAllUsers(ctx context.Context) ([]domain.User, error)
}

type userService struct {
  userRepo ports.UserRepository
}

func NewUserService(userRepo ports.UserRepository) UserService {
  return &userService{
    userRepo: userRepo,
  }
}

func (s *userService) CreateUser(ctx context.Context, user domain.User) (string, error) {
  // admin only
  return s.userRepo.SaveUser(ctx, user)
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
  // should be authorized
  // should be the same user or admin
  return s.userRepo.FindUserById(ctx, userID)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
  // admin only
  return s.userRepo.FindAllUsers(ctx)
}
