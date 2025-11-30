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

func Test_services_NewProjectService(t *testing.T) {
	projectRepo := mockrepository.NewMockProjectRepository()
	projectService := NewProjectService(projectRepo)
	require.NotNil(t, projectService, "NewProjectService() returned nil")
}

func Test_services_ProjectService_CreateProject(t *testing.T) {
	validDepartmentID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx     context.Context
		project domain.Project
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ProjectRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "create project successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					Name:         "Project Alpha",
					Description:  "A test project",
					Budget:       50000.00,
					DepartmentID: validDepartmentID,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projectRepo.On("SaveProject", mock.Anything, mock.AnythingOfType("domain.Project")).Return("new-project-id", nil)
				return projectRepo
			},
			wantErr: false,
		},
		{
			name: "create project without department",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					Name:        "Project Beta",
					Description: "A test project",
					Budget:      50000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projectRepo.On("SaveProject", mock.Anything, mock.AnythingOfType("domain.Project")).Return("new-project-id", nil)
				return projectRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				project: domain.Project{
					Name:        "Project Alpha",
					Description: "A test project",
					Budget:      50000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
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
				project: domain.Project{
					Name:        "Project Alpha",
					Description: "A test project",
					Budget:      50000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid project data - empty name",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					Name:        "",
					Description: "A test project",
					Budget:      50000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid project data - negative budget",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					Name:        "Project Alpha",
					Description: "A test project",
					Budget:      -1000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid department ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					Name:         "Project Alpha",
					Description:  "A test project",
					Budget:       50000.00,
					DepartmentID: "invalid-id",
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRepo := tt.getMockRepo()
			projectService := NewProjectService(projectRepo)

			result, err := projectService.CreateProject(tt.args.ctx, tt.args.project)

			if tt.wantErr {
				assert.Error(t, err, "CreateProject() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "CreateProject() should not return error")
				assert.NotEmpty(t, result, "CreateProject() should return project ID")
			}

			projectRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ProjectService_GetProjectByID(t *testing.T) {
	validProjectID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx       context.Context
		projectID string
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ProjectRepository
		want        domain.Project
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get project by ID successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				projectID: validProjectID,
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projectRepo.On("FindProjectById", mock.Anything, validProjectID).Return(domain.Project{
					ID:          validProjectID,
					Name:        "Project Alpha",
					Description: "A test project",
					Budget:      50000.00,
				}, nil)
				return projectRepo
			},
			want: domain.Project{
				ID:          validProjectID,
				Name:        "Project Alpha",
				Description: "A test project",
				Budget:      50000.00,
			},
			wantErr: false,
		},
		{
			name: "invalid project ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				projectID: "invalid-id",
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:       context.Background(),
				projectID: validProjectID,
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "project not found",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				projectID: validProjectID,
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projectRepo.On("FindProjectById", mock.Anything, validProjectID).
					Return(domain.Project{}, apperr.NewAppError(apperr.ErrNotFound, "project not found", nil))
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRepo := tt.getMockRepo()
			projectService := NewProjectService(projectRepo)

			result, err := projectService.GetProjectByID(tt.args.ctx, tt.args.projectID)

			if tt.wantErr {
				assert.Error(t, err, "GetProjectByID() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetProjectByID() should not return error")
				assert.Equal(t, tt.want, result, "GetProjectByID() returned incorrect project")
			}

			projectRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ProjectService_UpdateProject(t *testing.T) {
	validProjectID := "550e8400-e29b-41d4-a716-446655440000"
	validDepartmentID := "650e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx     context.Context
		project domain.Project
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.ProjectRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update project successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					ID:           validProjectID,
					Name:         "Updated Project",
					Description:  "Updated description",
					Budget:       75000.00,
					DepartmentID: validDepartmentID,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projectRepo.On("FindProjectById", mock.Anything, validProjectID).Return(domain.Project{
					ID:        validProjectID,
					Name:      "Project Alpha",
					Budget:    50000.00,
					CreatedAt: 1234567890,
				}, nil)
				projectRepo.On("UpdateProject", mock.Anything, mock.AnythingOfType("domain.Project")).Return(nil)
				return projectRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				project: domain.Project{
					ID:          validProjectID,
					Name:        "Updated Project",
					Description: "Updated description",
					Budget:      75000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
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
				project: domain.Project{
					ID:          validProjectID,
					Name:        "Updated Project",
					Description: "Updated description",
					Budget:      75000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid project ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					ID:          "invalid-id",
					Name:        "Updated Project",
					Description: "Updated description",
					Budget:      75000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid project data - empty name",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					ID:          validProjectID,
					Name:        "",
					Description: "Updated description",
					Budget:      75000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid department ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					ID:           validProjectID,
					Name:         "Updated Project",
					Description:  "Updated description",
					Budget:       75000.00,
					DepartmentID: "invalid-dept-id",
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "project not found",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				project: domain.Project{
					ID:          validProjectID,
					Name:        "Updated Project",
					Description: "Updated description",
					Budget:      75000.00,
				},
			},
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projectRepo.On("FindProjectById", mock.Anything, validProjectID).
					Return(domain.Project{}, apperr.NewAppError(apperr.ErrNotFound, "project not found", nil))
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRepo := tt.getMockRepo()
			projectService := NewProjectService(projectRepo)

			err := projectService.UpdateProject(tt.args.ctx, tt.args.project)

			if tt.wantErr {
				assert.Error(t, err, "UpdateProject() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateProject() should not return error")
			}

			projectRepo.AssertExpectations(t)
		})
	}
}

func Test_services_ProjectService_GetAllProjects(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		getMockRepo func() *mockrepository.ProjectRepository
		want        []domain.Project
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get all projects successfully",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "admin-id",
				Role:   domain.Admin,
			}),
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projects := []domain.Project{
					{ID: "project-1", Name: "Project Alpha", Budget: 50000.00},
					{ID: "project-2", Name: "Project Beta", Budget: 75000.00},
				}
				projectRepo.On("FindAllProjects", mock.Anything).Return(projects, nil)
				return projectRepo
			},
			want: []domain.Project{
				{ID: "project-1", Name: "Project Alpha", Budget: 50000.00},
				{ID: "project-2", Name: "Project Beta", Budget: 75000.00},
			},
			wantErr: false,
		},
		{
			name: "employee can also get all projects",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "employee-id",
				Role:   domain.Employee,
			}),
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				projects := []domain.Project{
					{ID: "project-1", Name: "Project Alpha", Budget: 50000.00},
				}
				projectRepo.On("FindAllProjects", mock.Anything).Return(projects, nil)
				return projectRepo
			},
			want: []domain.Project{
				{ID: "project-1", Name: "Project Alpha", Budget: 50000.00},
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			ctx:  context.Background(),
			getMockRepo: func() *mockrepository.ProjectRepository {
				projectRepo := mockrepository.NewMockProjectRepository()
				return projectRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRepo := tt.getMockRepo()
			projectService := NewProjectService(projectRepo)

			result, err := projectService.GetAllProjects(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetAllProjects() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAllProjects() should not return error")
				assert.Equal(t, tt.want, result, "GetAllProjects() returned incorrect projects")
			}

			projectRepo.AssertExpectations(t)
		})
	}
}
