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

func Test_mysql_DepartmentRepository_SaveDepartment(t *testing.T) {
	tests := []struct {
		name       string
		department domain.Department
		setupMock  func(sqlmock.Sqlmock, domain.Department)
		wantErr    bool
		errCode    apperr.AppErrorCode
	}{
		{
			name: "success - save department",
			department: domain.Department{
				ID:     "dept-1",
				Name:   "Engineering",
				Budget: 100000.50,
			},
			setupMock: func(mock sqlmock.Sqlmock, dept domain.Department) {
				mock.ExpectExec("INSERT INTO departments").
					WithArgs(dept.ID, dept.Name, dept.Budget).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "error - duplicate name",
			department: domain.Department{
				ID:     "dept-2",
				Name:   "Duplicate",
				Budget: 50000.00,
			},
			setupMock: func(mock sqlmock.Sqlmock, dept domain.Department) {
				mock.ExpectExec("INSERT INTO departments").
					WithArgs(dept.ID, dept.Name, dept.Budget).
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

			tt.setupMock(mock, tt.department)

			repo := NewDepartmentRepository(db)
			id, err := repo.SaveDepartment(context.Background(), tt.department)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.department.ID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_DepartmentRepository_UpdateDepartment(t *testing.T) {
	tests := []struct {
		name       string
		department domain.Department
		setupMock  func(sqlmock.Sqlmock, domain.Department)
		wantErr    bool
		errCode    apperr.AppErrorCode
	}{
		{
			name: "success - update department",
			department: domain.Department{
				ID:     "dept-1",
				Name:   "Engineering Updated",
				Budget: 150000.75,
			},
			setupMock: func(mock sqlmock.Sqlmock, dept domain.Department) {
				mock.ExpectExec("UPDATE departments").
					WithArgs(dept.Name, dept.Budget, dept.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error - department not found",
			department: domain.Department{
				ID:     "non-existent",
				Name:   "Ghost Department",
				Budget: 0,
			},
			setupMock: func(mock sqlmock.Sqlmock, dept domain.Department) {
				mock.ExpectExec("UPDATE departments").
					WithArgs(dept.Name, dept.Budget, dept.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
		{
			name: "error - duplicate name",
			department: domain.Department{
				ID:     "dept-1",
				Name:   "Duplicate Name",
				Budget: 100000,
			},
			setupMock: func(mock sqlmock.Sqlmock, dept domain.Department) {
				mock.ExpectExec("UPDATE departments").
					WithArgs(dept.Name, dept.Budget, dept.ID).
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

			tt.setupMock(mock, tt.department)

			repo := NewDepartmentRepository(db)
			err = repo.UpdateDepartment(context.Background(), tt.department)

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

func Test_mysql_DepartmentRepository_FindDepartmentById(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name         string
		departmentID string
		setupMock    func(sqlmock.Sqlmock, string)
		want         domain.Department
		wantErr      bool
		errCode      apperr.AppErrorCode
	}{
		{
			name:         "success - find department by id",
			departmentID: "dept-1",
			setupMock: func(mock sqlmock.Sqlmock, deptID string) {
				rows := sqlmock.NewRows([]string{"id", "name", "budget", "created_at", "updated_at"}).
					AddRow("dept-1", "Engineering", 100000.50, now, now)
				mock.ExpectQuery("SELECT (.+) FROM departments WHERE id = ?").
					WithArgs(deptID).
					WillReturnRows(rows)
			},
			want: domain.Department{
				ID:        "dept-1",
				Name:      "Engineering",
				Budget:    100000.50,
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name:         "error - department not found",
			departmentID: "non-existent",
			setupMock: func(mock sqlmock.Sqlmock, deptID string) {
				mock.ExpectQuery("SELECT (.+) FROM departments WHERE id = ?").
					WithArgs(deptID).
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

			tt.setupMock(mock, tt.departmentID)

			repo := NewDepartmentRepository(db)
			dept, err := repo.FindDepartmentById(context.Background(), tt.departmentID)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, dept)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_DepartmentRepository_FindAllDepartments(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		want      []domain.Department
		wantErr   bool
	}{
		{
			name: "success - find all departments",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "budget", "created_at", "updated_at"}).
					AddRow("dept-1", "Engineering", 100000.50, now, now).
					AddRow("dept-2", "Marketing", 75000.25, now, now)

				mock.ExpectQuery("SELECT (.+) FROM departments").WillReturnRows(rows)
			},
			want: []domain.Department{
				{ID: "dept-1", Name: "Engineering", Budget: 100000.50, CreatedAt: now, UpdatedAt: now},
				{ID: "dept-2", Name: "Marketing", Budget: 75000.25, CreatedAt: now, UpdatedAt: now},
			},
			wantErr: false,
		},
		{
			name: "success - empty list",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "budget", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT (.+) FROM departments").WillReturnRows(rows)
			},
			want:    []domain.Department{},
			wantErr: false,
		},
		{
			name: "error - query failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM departments").
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

			repo := NewDepartmentRepository(db)
			depts, err := repo.FindAllDepartments(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, depts)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
