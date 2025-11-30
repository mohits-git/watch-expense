package services

import (
	"context"
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockpasswordhasher "github.com/mohits-git/watch-expense/tests/mock_password_hasher"
	mockrepository "github.com/mohits-git/watch-expense/tests/mock_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_services_NewUserService(t *testing.T) {
	userRepo := mockrepository.NewMockUserRepository()
	projectRepo := mockrepository.NewMockProjectRepository()
	passwordHasher := &mockpasswordhasher.PasswordHasher{}
	userService := NewUserService(userRepo, projectRepo, passwordHasher)
	assert.NotNil(t, userService, "NewUserService() returned nil")
}

func Test_services_UserService_CreateUser(t *testing.T) {
	type args struct {
		ctx  context.Context
		user domain.User
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher)
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "create user successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				user: domain.User{
					Email:    "test@example.com",
					Name:     "Test User",
					Role:     domain.Employee,
					Password: "password123",
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				passwordHasher.On("HashPassword", "password123").Return("hashed_password", nil)
				userRepo.On("SaveUser", mock.Anything, mock.AnythingOfType("domain.User")).Return("new-user-id", nil)
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims in context",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					Email:    "test@example.com",
					Name:     "Test User",
					Role:     domain.Employee,
					Password: "password123",
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "unauthorized - user is not admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "employee-id",
					Role:   domain.Employee,
				}),
				user: domain.User{
					Email:    "test@example.com",
					Name:     "Test User",
					Role:     domain.Employee,
					Password: "password123",
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "invalid user data - empty email",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				user: domain.User{
					Email:    "",
					Name:     "Test User",
					Role:     domain.Employee,
					Password: "password123",
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid user data - empty name",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				user: domain.User{
					Email:    "test@example.com",
					Name:     "",
					Role:     domain.Employee,
					Password: "password123",
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid user data - empty role",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				user: domain.User{
					Email:    "test@example.com",
					Name:     "Test User",
					Role:     "",
					Password: "password123",
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, projectRepo, passwordHasher := tt.getMockRepo()
			userService := NewUserService(userRepo, projectRepo, passwordHasher)

			result, err := userService.CreateUser(tt.args.ctx, tt.args.user)

			if tt.wantErr {
				assert.Error(t, err, "CreateUser() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoErrorf(t, err, "CreateUser() error = %v", err)
				assert.NotEmpty(t, result, "CreateUser() should return user ID")
			}

			userRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func Test_services_UserService_UpdateUser(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	type args struct {
		ctx  context.Context
		user domain.User
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher)
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update user successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: userID,
					Role:   domain.Admin,
				}),
				user: domain.User{
					ID:    userID,
					Email: "updated@example.com",
					Name:  "Updated User",
					Role:  domain.Employee,
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("domain.User")).Return(nil)
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				user: domain.User{
					ID:    userID,
					Email: "updated@example.com",
					Name:  "Updated User",
					Role:  domain.Employee,
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "unauthorized - user is not admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: userID,
					Role:   domain.Employee,
				}),
				user: domain.User{
					ID:    userID,
					Email: "updated@example.com",
					Name:  "Updated User",
					Role:  domain.Employee,
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "invalid user data - empty ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: userID,
					Role:   domain.Admin,
				}),
				user: domain.User{
					ID:    "",
					Email: "updated@example.com",
					Name:  "Updated User",
					Role:  domain.Employee,
				},
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, projectRepo, passwordHasher := tt.getMockRepo()
			userService := NewUserService(userRepo, projectRepo, passwordHasher)

			err := userService.UpdateUser(tt.args.ctx, tt.args.user)

			if tt.wantErr {
				assert.Error(t, err, "UpdateUser() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateUser() should not return error")
			}

			userRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func Test_services_UserService_GetUserByID(t *testing.T) {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher)
		want        domain.User
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get user by ID successfully - admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: validUserID,
					Role:   domain.Admin,
				}),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:    validUserID,
					Email: "test@example.com",
					Name:  "Test User",
					Role:  domain.Employee,
				}, nil)
				return userRepo, projectRepo, passwordHasher
			},
			want: domain.User{
				ID:    validUserID,
				Email: "test@example.com",
				Name:  "Test User",
				Role:  domain.Employee,
			},
			wantErr: false,
		},
		{
			name: "get user by ID successfully - own user",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: validUserID,
					Role:   domain.Employee,
				}),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:    validUserID,
					Email: "test@example.com",
					Name:  "Test User",
					Role:  domain.Employee,
				}, nil)
				return userRepo, projectRepo, passwordHasher
			},
			want: domain.User{
				ID:    validUserID,
				Email: "test@example.com",
				Name:  "Test User",
				Role:  domain.Employee,
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				userID: "invalid-id",
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:    context.Background(),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - employee accessing other user",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "other-employee-id",
					Role:   domain.Employee,
				}),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, projectRepo, passwordHasher := tt.getMockRepo()
			userService := NewUserService(userRepo, projectRepo, passwordHasher)

			result, err := userService.GetUserByID(tt.args.ctx, tt.args.userID)

			if tt.wantErr {
				assert.Error(t, err, "GetUserByID() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetUserByID() should not return error")
				assert.Equal(t, tt.want, result, "GetUserByID() returned incorrect user")
			}

			userRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func Test_services_UserService_GetAllUsers(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		getMockRepo func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher)
		want        []domain.User
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get all users successfully",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "admin-id",
				Role:   domain.Admin,
			}),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				users := []domain.User{
					{ID: "user-1", Email: "user1@example.com", Name: "User 1", Role: domain.Employee},
					{ID: "user-2", Email: "user2@example.com", Name: "User 2", Role: domain.Employee},
				}
				userRepo.On("FindAllUsers", mock.Anything).Return(users, nil)
				return userRepo, projectRepo, passwordHasher
			},
			want: []domain.User{
				{ID: "user-1", Email: "user1@example.com", Name: "User 1", Role: domain.Employee},
				{ID: "user-2", Email: "user2@example.com", Name: "User 2", Role: domain.Employee},
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			ctx:  context.Background(),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "unauthorized - user is not admin",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "employee-id",
				Role:   domain.Employee,
			}),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, projectRepo, passwordHasher := tt.getMockRepo()
			userService := NewUserService(userRepo, projectRepo, passwordHasher)

			result, err := userService.GetAllUsers(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetAllUsers() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAllUsers() should not return error")
				assert.Equal(t, tt.want, result, "GetAllUsers() returned incorrect users")
			}

			userRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func Test_services_UserService_DeleteUser(t *testing.T) {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher)
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "delete user successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("DeleteUser", mock.Anything, validUserID).Return(nil)
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:    context.Background(),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "unauthorized - user is not admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "employee-id",
					Role:   domain.Employee,
				}),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - admin cannot delete self",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: validUserID,
					Role:   domain.Admin,
				}),
				userID: validUserID,
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid user ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				userID: "invalid-id",
			},
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, projectRepo, passwordHasher := tt.getMockRepo()
			userService := NewUserService(userRepo, projectRepo, passwordHasher)

			err := userService.DeleteUser(tt.args.ctx, tt.args.userID)

			if tt.wantErr {
				assert.Error(t, err, "DeleteUser() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "DeleteUser() should not return error")
			}

			userRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func Test_services_UserService_GetUserBudget(t *testing.T) {
	validUserID := "550e8400-e29b-41d4-a716-446655440000"
	validProjectID := "650e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name        string
		ctx         context.Context
		getMockRepo func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher)
		want        float64
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get user budget successfully",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:        validUserID,
					ProjectID: validProjectID,
				}, nil)
				projectRepo.On("FindProjectById", mock.Anything, validProjectID).Return(domain.Project{
					ID:     validProjectID,
					Budget: 50000.00,
				}, nil)
				return userRepo, projectRepo, passwordHasher
			},
			want:    50000.00,
			wantErr: false,
		},
		{
			name: "user has no project assigned",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:        validUserID,
					ProjectID: "",
				}, nil)
				return userRepo, projectRepo, passwordHasher
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "project not found - returns zero budget",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("FindUserById", mock.Anything, validUserID).Return(domain.User{
					ID:        validUserID,
					ProjectID: validProjectID,
				}, nil)
				projectRepo.On("FindProjectById", mock.Anything, validProjectID).
					Return(domain.Project{}, apperr.NewAppError(apperr.ErrNotFound, "project not found", nil))
				return userRepo, projectRepo, passwordHasher
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			ctx:  context.Background(),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "user not found error",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: validUserID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() (*mockrepository.UserRepository, *mockrepository.ProjectRepository, *mockpasswordhasher.PasswordHasher) {
				userRepo := mockrepository.NewMockUserRepository()
				projectRepo := mockrepository.NewMockProjectRepository()
				passwordHasher := &mockpasswordhasher.PasswordHasher{}
				userRepo.On("FindUserById", mock.Anything, validUserID).
					Return(domain.User{}, apperr.NewAppError(apperr.ErrNotFound, "user not found", nil))
				return userRepo, projectRepo, passwordHasher
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, projectRepo, passwordHasher := tt.getMockRepo()
			userService := NewUserService(userRepo, projectRepo, passwordHasher)

			result, err := userService.GetUserBudget(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetUserBudget() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetUserBudget() should not return error")
				assert.Equal(t, tt.want, result, "GetUserBudget() returned incorrect budget")
			}

			userRepo.AssertExpectations(t)
			projectRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}
