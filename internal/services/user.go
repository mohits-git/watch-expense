package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	"github.com/mohits-git/watch-expense/internal/utils/validator"
)

type UserService interface {
	CreateUser(ctx context.Context, user domain.User) (string, error)
	UpdateUser(ctx context.Context, user domain.User) error
	GetUserByID(ctx context.Context, userID string) (domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
  DeleteUser(ctx context.Context, userID string) error
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
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims.Role != domain.Admin {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "only admin can create users", nil)
	}
	if !validator.ValidateUserCreation(user) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid user data", nil)
	}
	user.ID = uuid.New().String()
	return s.userRepo.SaveUser(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, user domain.User) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrUnauthorized, "only admin can update users", nil)
	}
	if !validator.ValidateUserUpdate(user) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid user data", nil)
	}
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	if !validator.ValidateUUID(userID) {
		return domain.User{}, apperr.NewAppError(apperr.ErrInvalid, "invalid user ID", nil)
	}
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.User{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}
	if claims.Role != domain.Admin && claims.UserID != userID {
		return domain.User{}, apperr.NewAppError(apperr.ErrForbidden, "access denied", nil)
	}
	return s.userRepo.FindUserById(ctx, userID)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims.Role != domain.Admin {
		return nil, apperr.NewAppError(apperr.ErrUnauthorized, "only admin can access all users", nil)
	}
	return s.userRepo.FindAllUsers(ctx)
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
  claims, ok := authctx.UserClaimsFromCtx(ctx)
  if !ok || claims.Role != domain.Admin {
    return apperr.NewAppError(apperr.ErrUnauthorized, "only admin can delete users", nil)
  }
  if claims.UserID == userID {
    return apperr.NewAppError(apperr.ErrForbidden, "admin cannot delete self", nil)
  }
  if !validator.ValidateUUID(userID) {
    return apperr.NewAppError(apperr.ErrInvalid, "invalid user ID", nil)
  }
  return s.userRepo.DeleteUser(ctx, userID)
}
