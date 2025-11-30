package services

import (
	"context"
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockrepository "github.com/mohits-git/watch-expense/tests/mock_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_services_NewDepartmentService(t *testing.T) {
	departmentRepo := mockrepository.NewMockDepartmentRepository()
	departmentService := NewDepartmentService(departmentRepo)
	require.NotNil(t, departmentService, "NewDepartmentService() returned nil")
}

func Test_services_DepartmentService_CreateDepartment(t *testing.T) {
	type args struct {
		ctx        context.Context
		department domain.Department
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.DepartmentRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "create department successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					Name:   "Engineering",
					Budget: 100000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				departmentRepo.On("SaveDepartment", mock.Anything, mock.AnythingOfType("domain.Department")).Return("new-dept-id", nil)
				return departmentRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				department: domain.Department{
					Name:   "Engineering",
					Budget: 100000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is not admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "employee-id",
					Role:   domain.Employee,
				}),
				department: domain.Department{
					Name:   "Engineering",
					Budget: 100000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid department data - empty name",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					Name:   "",
					Budget: 100000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid department data - negative budget",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					Name:   "Engineering",
					Budget: -1000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			departmentRepo := tt.getMockRepo()
			departmentService := NewDepartmentService(departmentRepo)

			result, err := departmentService.CreateDepartment(tt.args.ctx, tt.args.department)

			if tt.wantErr {
				assert.Error(t, err, "CreateDepartment() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "CreateDepartment() should not return error")
				assert.NotEmpty(t, result, "CreateDepartment() should return department ID")
			}

			departmentRepo.AssertExpectations(t)
		})
	}
}

func Test_services_DepartmentService_GetDepartmentByID(t *testing.T) {
	validDepartmentID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx          context.Context
		departmentID string
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.DepartmentRepository
		want        domain.Department
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get department by ID successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				departmentID: validDepartmentID,
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				departmentRepo.On("FindDepartmentById", mock.Anything, validDepartmentID).Return(domain.Department{
					ID:     validDepartmentID,
					Name:   "Engineering",
					Budget: 100000.00,
				}, nil)
				return departmentRepo
			},
			want: domain.Department{
				ID:     validDepartmentID,
				Name:   "Engineering",
				Budget: 100000.00,
			},
			wantErr: false,
		},
		{
			name: "invalid department ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				departmentID: "invalid-id",
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:          context.Background(),
				departmentID: validDepartmentID,
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "department not found",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				departmentID: validDepartmentID,
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				departmentRepo.On("FindDepartmentById", mock.Anything, validDepartmentID).
					Return(domain.Department{}, apperr.NewAppError(apperr.ErrNotFound, "department not found", nil))
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			departmentRepo := tt.getMockRepo()
			departmentService := NewDepartmentService(departmentRepo)

			result, err := departmentService.GetDepartmentByID(tt.args.ctx, tt.args.departmentID)

			if tt.wantErr {
				assert.Error(t, err, "GetDepartmentByID() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetDepartmentByID() should not return error")
				assert.Equal(t, tt.want, result, "GetDepartmentByID() returned incorrect department")
			}

			departmentRepo.AssertExpectations(t)
		})
	}
}

func Test_services_DepartmentService_UpdateDepartment(t *testing.T) {
	validDepartmentID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx        context.Context
		department domain.Department
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.DepartmentRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update department successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					ID:     validDepartmentID,
					Name:   "Updated Engineering",
					Budget: 150000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				departmentRepo.On("FindDepartmentById", mock.Anything, validDepartmentID).Return(domain.Department{
					ID:        validDepartmentID,
					Name:      "Engineering",
					Budget:    100000.00,
					CreatedAt: 1234567890,
				}, nil)
				departmentRepo.On("UpdateDepartment", mock.Anything, mock.AnythingOfType("domain.Department")).Return(nil)
				return departmentRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				department: domain.Department{
					ID:     validDepartmentID,
					Name:   "Updated Engineering",
					Budget: 150000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is not admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "employee-id",
					Role:   domain.Employee,
				}),
				department: domain.Department{
					ID:     validDepartmentID,
					Name:   "Updated Engineering",
					Budget: 150000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid department ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					ID:     "invalid-id",
					Name:   "Updated Engineering",
					Budget: 150000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid department data - empty name",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					ID:     validDepartmentID,
					Name:   "",
					Budget: 150000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "department not found",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				department: domain.Department{
					ID:     validDepartmentID,
					Name:   "Updated Engineering",
					Budget: 150000.00,
				},
			},
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				departmentRepo.On("FindDepartmentById", mock.Anything, validDepartmentID).
					Return(domain.Department{}, apperr.NewAppError(apperr.ErrNotFound, "department not found", nil))
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			departmentRepo := tt.getMockRepo()
			departmentService := NewDepartmentService(departmentRepo)

			err := departmentService.UpdateDepartment(tt.args.ctx, tt.args.department)

			if tt.wantErr {
				assert.Error(t, err, "UpdateDepartment() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateDepartment() should not return error")
			}

			departmentRepo.AssertExpectations(t)
		})
	}
}

func Test_services_DepartmentService_GetAllDepartments(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		getMockRepo func() *mockrepository.DepartmentRepository
		want        []domain.Department
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get all departments successfully",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "admin-id",
				Role:   domain.Admin,
			}),
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				departments := []domain.Department{
					{ID: "dept-1", Name: "Engineering", Budget: 100000.00},
					{ID: "dept-2", Name: "Marketing", Budget: 50000.00},
				}
				departmentRepo.On("FindAllDepartments", mock.Anything).Return(departments, nil)
				return departmentRepo
			},
			want: []domain.Department{
				{ID: "dept-1", Name: "Engineering", Budget: 100000.00},
				{ID: "dept-2", Name: "Marketing", Budget: 50000.00},
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			ctx:  context.Background(),
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is not admin",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "employee-id",
				Role:   domain.Employee,
			}),
			getMockRepo: func() *mockrepository.DepartmentRepository {
				departmentRepo := mockrepository.NewMockDepartmentRepository()
				return departmentRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			departmentRepo := tt.getMockRepo()
			departmentService := NewDepartmentService(departmentRepo)

			result, err := departmentService.GetAllDepartments(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetAllDepartments() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAllDepartments() should not return error")
				assert.Equal(t, tt.want, result, "GetAllDepartments() returned incorrect departments")
			}

			departmentRepo.AssertExpectations(t)
		})
	}
}
