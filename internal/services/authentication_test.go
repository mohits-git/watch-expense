package services

import (
	"context"
	"errors"
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockpasswordhasher "github.com/mohits-git/watch-expense/tests/mock_password_hasher"
	mockrepository "github.com/mohits-git/watch-expense/tests/mock_repository"
	mocktokenprovider "github.com/mohits-git/watch-expense/tests/mock_token_provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_services_NewAuthenticationService(t *testing.T) {
	mockUserRepo := mockrepository.NewMockUserRepository()
	mockTokenProvider := &mocktokenprovider.TokenProvider{}
	mockPasswordHasher := &mockpasswordhasher.PasswordHasher{}

	service := NewAuthenticationService(mockUserRepo, mockTokenProvider, mockPasswordHasher)
	require.NotNil(t, service, "NewAuthenticationService() returned nil")
}

func Test_services_AuthenticationService_Login(t *testing.T) {
	type args struct {
		ctx      context.Context
		email    string
		password string
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(*mockrepository.UserRepository, *mocktokenprovider.TokenProvider, *mockpasswordhasher.PasswordHasher)
		wantToken  string
		wantErr    bool
		errCode    apperr.AppErrorCode
	}{
		{
			name: "login successfully",
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "password123",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				user := domain.User{
					ID:       "user-id-123",
					Name:     "Test User",
					Email:    "test@example.com",
					Password: "hashed-password",
					Role:     domain.Employee,
				}
				userRepo.On("FindUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
				passwordHasher.On("ComparePassword", "hashed-password", "password123").Return(true, nil)
				tokenProvider.On("GenerateToken", mock.AnythingOfType("authctx.UserClaims")).Return("valid.jwt.token", nil)
			},
			wantToken: "valid.jwt.token",
			wantErr:   false,
		},
		{
			name: "user not found",
			args: args{
				ctx:      context.Background(),
				email:    "nonexistent@example.com",
				password: "password123",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				userRepo.On("FindUserByEmail", mock.Anything, "nonexistent@example.com").
					Return(domain.User{}, apperr.NewAppError(apperr.ErrNotFound, "user not found", nil))
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
		{
			name: "password mismatch",
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "wrongpassword",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				user := domain.User{
					ID:       "user-id-123",
					Email:    "test@example.com",
					Password: "hashed-password",
					Role:     domain.Employee,
				}
				userRepo.On("FindUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
				passwordHasher.On("ComparePassword", "hashed-password", "wrongpassword").Return(false, nil)
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "password hasher error",
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "password123",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				user := domain.User{
					ID:       "user-id-123",
					Email:    "test@example.com",
					Password: "hashed-password",
					Role:     domain.Employee,
				}
				userRepo.On("FindUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
				passwordHasher.On("ComparePassword", "hashed-password", "password123").
					Return(false, errors.New("hasher error"))
			},
			wantErr: true,
		},
		{
			name: "token generation error",
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "password123",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				user := domain.User{
					ID:       "user-id-123",
					Email:    "test@example.com",
					Password: "hashed-password",
					Role:     domain.Employee,
				}
				userRepo.On("FindUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
				passwordHasher.On("ComparePassword", "hashed-password", "password123").Return(true, nil)
				tokenProvider.On("GenerateToken", mock.AnythingOfType("authctx.UserClaims")).
					Return("", errors.New("token generation failed"))
			},
			wantErr: true,
		},
		{
			name: "admin user login successfully",
			args: args{
				ctx:      context.Background(),
				email:    "admin@example.com",
				password: "adminpass",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				user := domain.User{
					ID:       "admin-id-456",
					Name:     "Admin User",
					Email:    "admin@example.com",
					Password: "hashed-admin-password",
					Role:     domain.Admin,
				}
				userRepo.On("FindUserByEmail", mock.Anything, "admin@example.com").Return(user, nil)
				passwordHasher.On("ComparePassword", "hashed-admin-password", "adminpass").Return(true, nil)
				tokenProvider.On("GenerateToken", mock.AnythingOfType("authctx.UserClaims")).Return("admin.jwt.token", nil)
			},
			wantToken: "admin.jwt.token",
			wantErr:   false,
		},
		{
			name: "repository error",
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				password: "password123",
			},
			setupMocks: func(userRepo *mockrepository.UserRepository, tokenProvider *mocktokenprovider.TokenProvider, passwordHasher *mockpasswordhasher.PasswordHasher) {
				userRepo.On("FindUserByEmail", mock.Anything, "test@example.com").
					Return(domain.User{}, apperr.NewAppError(apperr.ErrInternal, "database error", nil))
			},
			wantErr: true,
			errCode: apperr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mockrepository.NewMockUserRepository()
			tokenProvider := &mocktokenprovider.TokenProvider{}
			passwordHasher := &mockpasswordhasher.PasswordHasher{}

			tt.setupMocks(userRepo, tokenProvider, passwordHasher)

			service := NewAuthenticationService(userRepo, tokenProvider, passwordHasher)
			token, err := service.Login(tt.args.ctx, tt.args.email, tt.args.password)

			if tt.wantErr {
				assert.Error(t, err, "Login() should return error")
				if tt.errCode != 0 {
					appErr, ok := err.(*apperr.AppError)
					assert.True(t, ok, "Error should be of type *apperr.AppError")
					assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
				}
				assert.Empty(t, token, "Login() should not return token on error")
			} else {
				assert.NoError(t, err, "Login() should not return error")
				assert.Equal(t, tt.wantToken, token, "Login() returned incorrect token")
			}

			userRepo.AssertExpectations(t)
			tokenProvider.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func Test_services_AuthenticationService_Logout(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "logout with valid token",
			token:   "valid.jwt.token",
			wantErr: false,
		},
		{
			name:    "logout with empty token",
			token:   "",
			wantErr: false,
		},
		{
			name:    "logout with invalid token format",
			token:   "invalid-token",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mockrepository.NewMockUserRepository()
			tokenProvider := &mocktokenprovider.TokenProvider{}
			passwordHasher := &mockpasswordhasher.PasswordHasher{}

			service := NewAuthenticationService(userRepo, tokenProvider, passwordHasher)
			err := service.Logout(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err, "Logout() should return error")
			} else {
				assert.NoError(t, err, "Logout() should not return error")
			}
		})
	}
}

func Test_services_AuthenticationService_GetCurrentUser(t *testing.T) {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name       string
		ctx        context.Context
		setupMocks func(*mockrepository.UserRepository)
		want       domain.User
		wantErr    bool
		errCode    apperr.AppErrorCode
	}{
		{
			name: "get current user successfully",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Name:   "Test User",
				Email:  "test@example.com",
				Role:   domain.Employee,
			}),
			setupMocks: func(userRepo *mockrepository.UserRepository) {
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:    validUserID,
					Name:  "Test User",
					Email: "test@example.com",
					Role:  domain.Employee,
				}, nil)
			},
			want: domain.User{
				ID:    validUserID,
				Name:  "Test User",
				Email: "test@example.com",
				Role:  domain.Employee,
			},
			wantErr: false,
		},
		{
			name: "get current user - admin",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Name:   "Admin User",
				Email:  "admin@example.com",
				Role:   domain.Admin,
			}),
			setupMocks: func(userRepo *mockrepository.UserRepository) {
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:    validUserID,
					Name:  "Admin User",
					Email: "admin@example.com",
					Role:  domain.Admin,
				}, nil)
			},
			want: domain.User{
				ID:    validUserID,
				Name:  "Admin User",
				Email: "admin@example.com",
				Role:  domain.Admin,
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims in context",
			ctx:  context.Background(),
			setupMocks: func(userRepo *mockrepository.UserRepository) {
				// No mocks needed
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "user not found in repository",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Name:   "Test User",
				Email:  "test@example.com",
				Role:   domain.Employee,
			}),
			setupMocks: func(userRepo *mockrepository.UserRepository) {
				userRepo.On("FindUserById", mock.Anything, validUserID).
					Return(domain.User{}, apperr.NewAppError(apperr.ErrNotFound, "user not found", nil))
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
		{
			name: "repository error",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Name:   "Test User",
				Email:  "test@example.com",
				Role:   domain.Employee,
			}),
			setupMocks: func(userRepo *mockrepository.UserRepository) {
				userRepo.On("FindUserById", mock.Anything, validUserID).
					Return(domain.User{}, apperr.NewAppError(apperr.ErrInternal, "database error", nil))
			},
			wantErr: true,
			errCode: apperr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mockrepository.NewMockUserRepository()
			tokenProvider := &mocktokenprovider.TokenProvider{}
			passwordHasher := &mockpasswordhasher.PasswordHasher{}

			tt.setupMocks(userRepo)

			service := NewAuthenticationService(userRepo, tokenProvider, passwordHasher)
			user, err := service.GetCurrentUser(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetCurrentUser() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetCurrentUser() should not return error")
				assert.Equal(t, tt.want, user, "GetCurrentUser() returned incorrect user")
			}

			userRepo.AssertExpectations(t)
		})
	}
}
