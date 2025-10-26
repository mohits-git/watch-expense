package services

import (
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockpasswordhasher "github.com/mohits-git/watch-expense/tests/mock_password_hasher"
	mockrepository "github.com/mohits-git/watch-expense/tests/mock_repository"
	mocktokenprovider "github.com/mohits-git/watch-expense/tests/mock_token_provider"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_services_AuthenticationService_NewAuthenticationService(t *testing.T) {
	mockUserRepo := mockrepository.UserRepository{}
	mockTokenProvider := mocktokenprovider.TokenProvider{}
	mockPasswordHasher := mockpasswordhasher.PasswordHasher{}

	service := NewAuthenticationService(&mockUserRepo, &mockTokenProvider, &mockPasswordHasher)
	require.NotNil(t, service)
}

func Test_services_AuthenticationService_Login(t *testing.T) {
	mockUserRepo := mockrepository.UserRepository{}
	mockTokenProvider := mocktokenprovider.TokenProvider{}
	mockPasswordHasher := mockpasswordhasher.PasswordHasher{}

	service := NewAuthenticationService(&mockUserRepo, &mockTokenProvider, &mockPasswordHasher)

	email := "test@example.com"
	password := "password123"
	hashedPassword := "hashedpassword123"
	userID := "1"
	userRole := "customer"
	expectedToken := "valid.jwt.token"
	user := domain.User{ID: userID, Email: email, Password: hashedPassword, Role: domain.UserRole(userRole)}
	userClaims := authctx.NewUserClaims(userID, user.Name, email, domain.UserRole(userRole))
	// Mocking the user repository to return a user
	mockUserRepo.On("FindUserByEmail", mock.Anything, email).
		Return(user, nil)
	// Mocking the password hasher to return a successful match
	mockPasswordHasher.On("ComparePassword", hashedPassword, password).
		Return(true, nil)
	// Mocking the token provider to return a valid token
	mockTokenProvider.On("GenerateToken", userClaims).
		Return(expectedToken, nil)
	token, err := service.Login(t.Context(), email, password)
	require.NoError(t, err)
	require.Equal(t, expectedToken, token)
	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
	mockTokenProvider.AssertExpectations(t)
}

func Test_services_AuthenticationService_Login_when_user_not_found(t *testing.T) {
	mockUserRepo := mockrepository.UserRepository{}
	mockTokenProvider := mocktokenprovider.TokenProvider{}
	mockPasswordHasher := mockpasswordhasher.PasswordHasher{}

	service := NewAuthenticationService(&mockUserRepo, &mockTokenProvider, &mockPasswordHasher)

	email := "test@example.com"
	password := "password123"
	expectedErr := apperr.NewAppError(apperr.ErrNotFound, "user not found", nil)
	// mock
	mockUserRepo.On("FindUserByEmail", mock.Anything, email).
		Return(domain.User{}, expectedErr)
	token, err := service.Login(t.Context(), email, password)
	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, token)
	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertNotCalled(t, "ComparePassword", mock.Anything, mock.Anything)
	mockTokenProvider.AssertNotCalled(t, "GenerateToken", mock.Anything)
}

func Test_services_AuthenticationService_Login_when_password_mismatch(t *testing.T) {
	mockUserRepo := mockrepository.UserRepository{}
	mockTokenProvider := mocktokenprovider.TokenProvider{}
	mockPasswordHasher := mockpasswordhasher.PasswordHasher{}

	service := NewAuthenticationService(&mockUserRepo, &mockTokenProvider, &mockPasswordHasher)

	email := "test@example.com"
	password := "password123"
	hashedPassword := "hashedpassword123"
	userID := "1"
	userRole := "customer"
	user := domain.User{ID: userID, Email: email, Password: hashedPassword, Role: domain.UserRole(userRole)}
	expectedErr := apperr.NewAppError(apperr.ErrUnauthorized, "invalid email or password", nil)

	mockUserRepo.On("FindUserByEmail", mock.Anything, email).
		Return(user, nil)
	mockPasswordHasher.On("ComparePassword", hashedPassword, password).
		Return(false, nil)

	token, err := service.Login(t.Context(), email, password)
	apperr, ok := err.(*apperr.AppError)
	require.True(t, ok)
	require.Equal(t, expectedErr.Code, apperr.Code)
	require.Empty(t, token)
	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
	mockTokenProvider.AssertNotCalled(t, "GenerateToken", mock.Anything)
}

func Test_services_AuthenticationService_Logout(t *testing.T) {
	mockUserRepo := mockrepository.UserRepository{}
	mockTokenProvider := mocktokenprovider.TokenProvider{}
	mockPasswordHasher := mockpasswordhasher.PasswordHasher{}

	service := NewAuthenticationService(&mockUserRepo, &mockTokenProvider, &mockPasswordHasher)

	err := service.Logout(t.Context(), "some.jwt.token")
	require.NoError(t, err)
}
