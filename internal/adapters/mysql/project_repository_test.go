package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mysql_ProjectRepository_SaveProject(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.Add(30 * 24 * time.Hour)

	tests := []struct {
		name      string
		project   domain.Project
		setupMock func(sqlmock.Sqlmock, domain.Project)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - save project with department",
			project: domain.Project{
				ID:           "project-1",
				Name:         "New Website",
				Description:  "Build new company website",
				Budget:       50000.00,
				StartDate:    startDate.UnixMilli(),
				EndDate:      endDate.UnixMilli(),
				DepartmentID: "dept-1",
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("INSERT INTO projects").
					WithArgs(project.ID, project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{String: project.DepartmentID, Valid: true}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "success - save project without department",
			project: domain.Project{
				ID:          "project-2",
				Name:        "Internal Tool",
				Description: "Build internal tool",
				Budget:      25000.00,
				StartDate:   startDate.UnixMilli(),
				EndDate:     endDate.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("INSERT INTO projects").
					WithArgs(project.ID, project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "error - duplicate name",
			project: domain.Project{
				ID:          "project-3",
				Name:        "Duplicate",
				Description: "Duplicate project",
				Budget:      10000.00,
				StartDate:   startDate.UnixMilli(),
				EndDate:     endDate.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("INSERT INTO projects").
					WithArgs(project.ID, project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{}).
					WillReturnError(&mysql.MySQLError{Number: 1062})
			},
			wantErr: true,
			errCode: apperr.ErrConflict,
		},
		{
			name: "error - invalid department foreign key",
			project: domain.Project{
				ID:           "project-4",
				Name:         "Invalid FK Project",
				Description:  "Project with invalid department",
				Budget:       10000.00,
				StartDate:    startDate.UnixMilli(),
				EndDate:      endDate.UnixMilli(),
				DepartmentID: "invalid-dept",
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("INSERT INTO projects").
					WithArgs(project.ID, project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{String: project.DepartmentID, Valid: true}).
					WillReturnError(&mysql.MySQLError{Number: 1452})
			},
			wantErr: true,
			errCode: apperr.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock, tt.project)

			repo := NewProjectRepository(db)
			id, err := repo.SaveProject(context.Background(), tt.project)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.project.ID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_ProjectRepository_UpdateProject(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.Add(30 * 24 * time.Hour)

	tests := []struct {
		name      string
		project   domain.Project
		setupMock func(sqlmock.Sqlmock, domain.Project)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - update project",
			project: domain.Project{
				ID:           "project-1",
				Name:         "Updated Website",
				Description:  "Updated description",
				Budget:       75000.00,
				StartDate:    startDate.UnixMilli(),
				EndDate:      endDate.UnixMilli(),
				DepartmentID: "dept-2",
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("UPDATE projects").
					WithArgs(project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{String: project.DepartmentID, Valid: true}, project.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error - project not found",
			project: domain.Project{
				ID:          "non-existent",
				Name:        "Ghost Project",
				Description: "Does not exist",
				Budget:      0,
				StartDate:   startDate.UnixMilli(),
				EndDate:     endDate.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("UPDATE projects").
					WithArgs(project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{}, project.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
		{
			name: "error - duplicate name",
			project: domain.Project{
				ID:          "project-1",
				Name:        "Duplicate Name",
				Description: "Duplicate",
				Budget:      50000,
				StartDate:   startDate.UnixMilli(),
				EndDate:     endDate.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, project domain.Project) {
				mock.ExpectExec("UPDATE projects").
					WithArgs(project.Name, project.Description, project.Budget,
						time.UnixMilli(project.StartDate), time.UnixMilli(project.EndDate),
						sql.NullString{}, project.ID).
					WillReturnError(&mysql.MySQLError{Number: 1062})
			},
			wantErr: true,
			errCode: apperr.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock, tt.project)

			repo := NewProjectRepository(db)
			err = repo.UpdateProject(context.Background(), tt.project)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_ProjectRepository_FindProjectById(t *testing.T) {
	now := time.Now().Unix()
	startDate := time.Now().Unix()
	endDate := time.Now().Add(30 * 24 * time.Hour).Unix()

	tests := []struct {
		name      string
		projectID string
		setupMock func(sqlmock.Sqlmock, string)
		want      domain.Project
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name:      "success - find project with department",
			projectID: "project-1",
			setupMock: func(mock sqlmock.Sqlmock, projectID string) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "budget", "start_date", "end_date",
					"department_id", "created_at", "updated_at",
				}).AddRow(
					"project-1", "New Website", "Build website", 50000.00, startDate, endDate,
					sql.NullString{String: "dept-1", Valid: true}, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM projects WHERE id = ?").
					WithArgs(projectID).
					WillReturnRows(rows)
			},
			want: domain.Project{
				ID:           "project-1",
				Name:         "New Website",
				Description:  "Build website",
				Budget:       50000.00,
				StartDate:    startDate,
				EndDate:      endDate,
				DepartmentID: "dept-1",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name:      "success - find project without department",
			projectID: "project-2",
			setupMock: func(mock sqlmock.Sqlmock, projectID string) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "budget", "start_date", "end_date",
					"department_id", "created_at", "updated_at",
				}).AddRow(
					"project-2", "Internal Tool", "Build tool", 25000.00, startDate, endDate,
					sql.NullString{}, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM projects WHERE id = ?").
					WithArgs(projectID).
					WillReturnRows(rows)
			},
			want: domain.Project{
				ID:          "project-2",
				Name:        "Internal Tool",
				Description: "Build tool",
				Budget:      25000.00,
				StartDate:   startDate,
				EndDate:     endDate,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: false,
		},
		{
			name:      "error - project not found",
			projectID: "non-existent",
			setupMock: func(mock sqlmock.Sqlmock, projectID string) {
				mock.ExpectQuery("SELECT (.+) FROM projects WHERE id = ?").
					WithArgs(projectID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock, tt.projectID)

			repo := NewProjectRepository(db)
			project, err := repo.FindProjectById(context.Background(), tt.projectID)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, project)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_ProjectRepository_FindAllProjects(t *testing.T) {
	now := time.Now().Unix()
	startDate := time.Now().Unix()
	endDate := time.Now().Add(30 * 24 * time.Hour).Unix()

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		want      []domain.Project
		wantErr   bool
	}{
		{
			name: "success - find all projects",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "budget", "start_date", "end_date",
					"department_id", "created_at", "updated_at",
				}).
					AddRow("project-1", "New Website", "Build website", 50000.00, startDate, endDate,
						sql.NullString{String: "dept-1", Valid: true}, now, now).
					AddRow("project-2", "Internal Tool", "Build tool", 25000.00, startDate, endDate,
						sql.NullString{}, now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").WillReturnRows(rows)
			},
			want: []domain.Project{
				{
					ID: "project-1", Name: "New Website", Description: "Build website", Budget: 50000.00,
					StartDate: startDate, EndDate: endDate, DepartmentID: "dept-1",
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "project-2", Name: "Internal Tool", Description: "Build tool", Budget: 25000.00,
					StartDate: startDate, EndDate: endDate, CreatedAt: now, UpdatedAt: now,
				},
			},
			wantErr: false,
		},
		{
			name: "success - empty list",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "budget", "start_date", "end_date",
					"department_id", "created_at", "updated_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM projects").WillReturnRows(rows)
			},
			want:    []domain.Project{},
			wantErr: false,
		},
		{
			name: "error - query failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM projects").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewProjectRepository(db)
			projects, err := repo.FindAllProjects(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, projects)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
