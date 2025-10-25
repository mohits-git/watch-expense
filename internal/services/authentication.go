package services

import (
	"context"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

type AuthenticationService interface {
	Login(ctx context.Context, email, password string) (token string, err error)
	Logout(ctx context.Context, token string) error
	GetCurrentUser(ctx context.Context) (user domain.User, err error)
}

type authenticationService struct {
	userRepo       ports.UserRepository
	tokenProvider  ports.TokenProvider
	passwordHasher ports.PasswordHasher
}

func NewAuthenticationService(
	userRepo ports.UserRepository,
	tokenProvider ports.TokenProvider,
	passwordHasher ports.PasswordHasher,
) AuthenticationService {
	return &authenticationService{
		userRepo:       userRepo,
		tokenProvider:  tokenProvider,
		passwordHasher: passwordHasher,
	}
}

func (s *authenticationService) Login(ctx context.Context, email, password string) (token string, err error) {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	match, err := s.passwordHasher.ComparePassword(user.Password, password)
	if err != nil {
		return "", err
	}
	if !match {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "invalid email or password", nil)
	}

	claims := authctx.NewUserClaims(user.ID, user.Role)
	token, err = s.tokenProvider.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authenticationService) Logout(ctx context.Context, token string) error {
	// no op logout for stateless JWT
	// if we add blacklisting, we can implement it here
	// or refresh token mechanism
	return nil
}

func (s *authenticationService) GetCurrentUser(ctx context.Context) (user domain.User, err error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.User{}, apperr.NewAppError(apperr.ErrUnauthorized, "no authentication claims found", nil)
	}

	user, err = s.userRepo.FindUserById(ctx, claims.UserID)
	if err != nil {
		return
	}
	return user, nil
}
