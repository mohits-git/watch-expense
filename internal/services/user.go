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
	GetUserBudget(ctx context.Context) (float64, error)
}

type userService struct {
	userRepo       ports.UserRepository
	projectRepo    ports.ProjectRepository
	passwordHasher ports.PasswordHasher
}

func NewUserService(userRepo ports.UserRepository, projectRepo ports.ProjectRepository, passwordHasher ports.PasswordHasher) UserService {
	return &userService{
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		passwordHasher: passwordHasher,
	}
}

func (s *userService) CreateUser(ctx context.Context, user domain.User) (string, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}
	if claims.Role != domain.Admin {
		return "", apperr.NewAppError(apperr.ErrForbidden, "only admin can create users", nil)
	}
	if !validator.ValidateUserCreation(user) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid user data", nil)
	}
	hashedPassword, err := s.passwordHasher.HashPassword(user.Password)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to hash password", err)
	}
	user.Password = hashedPassword
	user.ID = uuid.New().String()
	return s.userRepo.SaveUser(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, user domain.User) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}
	if claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrForbidden, "only admin can update users", nil)
	}
	if !validator.ValidateUserUpdate(user) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid user data", nil)
	}
	if user.Password != "" {
		hashedPassword, err := s.passwordHasher.HashPassword(user.Password)
		if err != nil {
			return apperr.NewAppError(apperr.ErrInternal, "failed to hash password", err)
		}
		user.Password = hashedPassword
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
	if !ok {
		return nil, apperr.NewAppError(apperr.ErrUnauthorized, "only admin can access all users", nil)
	}
	if claims.Role != domain.Admin {
		return nil, apperr.NewAppError(apperr.ErrForbidden, "forbidden", nil)
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

func (s *userService) GetUserBudget(ctx context.Context) (float64, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return 0, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	user, err := s.userRepo.FindUserById(ctx, claims.UserID)
	if err != nil {
		return 0, err
	}

	if user.ProjectID == "" {
		return 0, nil
	}

	project, err := s.projectRepo.FindProjectById(ctx, user.ProjectID)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			return 0, nil
		}
		return 0, err
	}

	return project.Budget, nil
}
