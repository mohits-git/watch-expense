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

func Test_mysql_UserRepository_SaveUser(t *testing.T) {
	tests := []struct {
		name      string
		user      domain.User
		setupMock func(sqlmock.Sqlmock, domain.User)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - save user with all fields",
			user: domain.User{
				ID:           "user-1",
				EmployeeId:   "EMP001",
				Name:         "John Doe",
				Email:        "john@example.com",
				Password:     "hashedpassword",
				Role:         "employee",
				ProjectID:    "project-1",
				DepartmentID: "dept-1",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(user.ID, user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{String: user.ProjectID, Valid: true},
						sql.NullString{String: user.DepartmentID, Valid: true}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "success - save user without optional fields",
			user: domain.User{
				ID:         "user-2",
				EmployeeId: "EMP002",
				Name:       "Jane Smith",
				Email:      "jane@example.com",
				Password:   "hashedpassword",
				Role:       "manager",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(user.ID, user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{}, sql.NullString{}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "error - duplicate email",
			user: domain.User{
				ID:         "user-3",
				EmployeeId: "EMP003",
				Name:       "Duplicate User",
				Email:      "duplicate@example.com",
				Password:   "hashedpassword",
				Role:       "employee",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(user.ID, user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{}, sql.NullString{}).
					WillReturnError(&mysql.MySQLError{Number: 1062})
			},
			wantErr: true,
			errCode: apperr.ErrConflict,
		},
		{
			name: "error - foreign key constraint",
			user: domain.User{
				ID:           "user-4",
				EmployeeId:   "EMP004",
				Name:         "Invalid FK User",
				Email:        "invalid@example.com",
				Password:     "hashedpassword",
				Role:         "employee",
				DepartmentID: "invalid-dept",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(user.ID, user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{}, sql.NullString{String: user.DepartmentID, Valid: true}).
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

			tt.setupMock(mock, tt.user)

			repo := NewUserRepository(db)
			id, err := repo.SaveUser(context.Background(), tt.user)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.user.ID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_UserRepository_UpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		user      domain.User
		setupMock func(sqlmock.Sqlmock, domain.User)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - update user",
			user: domain.User{
				ID:           "user-1",
				EmployeeId:   "EMP001",
				Name:         "John Updated",
				Email:        "john.updated@example.com",
				Password:     "newhashedpassword",
				Role:         "manager",
				ProjectID:    "project-2",
				DepartmentID: "dept-2",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("UPDATE users").
					WithArgs(user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{String: user.ProjectID, Valid: true},
						sql.NullString{String: user.DepartmentID, Valid: true}, user.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error - user not found",
			user: domain.User{
				ID:         "non-existent",
				EmployeeId: "EMP999",
				Name:       "Ghost User",
				Email:      "ghost@example.com",
				Password:   "password",
				Role:       "employee",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("UPDATE users").
					WithArgs(user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{}, sql.NullString{}, user.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
		{
			name: "error - duplicate email",
			user: domain.User{
				ID:         "user-1",
				EmployeeId: "EMP001",
				Name:       "John Doe",
				Email:      "duplicate@example.com",
				Password:   "password",
				Role:       "employee",
			},
			setupMock: func(mock sqlmock.Sqlmock, user domain.User) {
				mock.ExpectExec("UPDATE users").
					WithArgs(user.EmployeeId, user.Name, user.Email, user.Password, user.Role,
						sql.NullString{}, sql.NullString{}, user.ID).
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

			tt.setupMock(mock, tt.user)

			repo := NewUserRepository(db)
			err = repo.UpdateUser(context.Background(), tt.user)

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

func Test_mysql_UserRepository_FindUserById(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		userID    string
		setupMock func(sqlmock.Sqlmock, string)
		want      domain.User
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name:   "success - find user with all fields",
			userID: "user-1",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{
					"id", "employee_id", "name", "email", "password_hash", "role",
					"project_id", "department_id", "created_at", "updated_at",
				}).AddRow(
					"user-1", "EMP001", "John Doe", "john@example.com", "hashedpassword", "employee",
					sql.NullString{String: "project-1", Valid: true},
					sql.NullString{String: "dept-1", Valid: true},
					now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
					WithArgs("user-1").
					WillReturnRows(rows)
			},
			want: domain.User{
				ID:           "user-1",
				EmployeeId:   "EMP001",
				Name:         "John Doe",
				Email:        "john@example.com",
				Password:     "hashedpassword",
				Role:         "employee",
				ProjectID:    "project-1",
				DepartmentID: "dept-1",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name:   "success - find user without optional fields",
			userID: "user-2",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{
					"id", "employee_id", "name", "email", "password_hash", "role",
					"project_id", "department_id", "created_at", "updated_at",
				}).AddRow(
					"user-2", "EMP002", "Jane Smith", "jane@example.com", "hashedpassword", "manager",
					sql.NullString{}, sql.NullString{}, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
					WithArgs("user-2").
					WillReturnRows(rows)
			},
			want: domain.User{
				ID:         "user-2",
				EmployeeId: "EMP002",
				Name:       "Jane Smith",
				Email:      "jane@example.com",
				Password:   "hashedpassword",
				Role:       "manager",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			wantErr: false,
		},
		{
			name:   "error - user not found",
			userID: "non-existent",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
					WithArgs("non-existent").
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

			tt.setupMock(mock, tt.userID)

			repo := NewUserRepository(db)
			user, err := repo.FindUserById(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_UserRepository_FindUserByEmail(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		email     string
		setupMock func(sqlmock.Sqlmock, string)
		want      domain.User
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name:  "success - find user by email",
			email: "john@example.com",
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				rows := sqlmock.NewRows([]string{
					"id", "employee_id", "name", "email", "password_hash", "role",
					"project_id", "department_id", "created_at", "updated_at",
				}).AddRow(
					"user-1", "EMP001", "John Doe", "john@example.com", "hashedpassword", "employee",
					sql.NullString{String: "project-1", Valid: true},
					sql.NullString{String: "dept-1", Valid: true},
					now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
					WithArgs(email).
					WillReturnRows(rows)
			},
			want: domain.User{
				ID:           "user-1",
				EmployeeId:   "EMP001",
				Name:         "John Doe",
				Email:        "john@example.com",
				Password:     "hashedpassword",
				Role:         "employee",
				ProjectID:    "project-1",
				DepartmentID: "dept-1",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name:  "error - user not found",
			email: "nonexistent@example.com",
			setupMock: func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
					WithArgs(email).
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

			tt.setupMock(mock, tt.email)

			repo := NewUserRepository(db)
			user, err := repo.FindUserByEmail(context.Background(), tt.email)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_UserRepository_FindAllUsers(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		want      []domain.User
		wantErr   bool
	}{
		{
			name: "success - find all users",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "employee_id", "name", "email", "password_hash", "role",
					"project_id", "department_id", "created_at", "updated_at",
				}).
					AddRow("user-1", "EMP001", "John Doe", "john@example.com", "pass1", "employee",
						sql.NullString{String: "project-1", Valid: true},
						sql.NullString{String: "dept-1", Valid: true}, now, now).
					AddRow("user-2", "EMP002", "Jane Smith", "jane@example.com", "pass2", "manager",
						sql.NullString{}, sql.NullString{}, now, now)

				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			want: []domain.User{
				{
					ID: "user-1", EmployeeId: "EMP001", Name: "John Doe", Email: "john@example.com",
					Password: "pass1", Role: "employee", ProjectID: "project-1", DepartmentID: "dept-1",
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "user-2", EmployeeId: "EMP002", Name: "Jane Smith", Email: "jane@example.com",
					Password: "pass2", Role: "manager", CreatedAt: now, UpdatedAt: now,
				},
			},
			wantErr: false,
		},
		{
			name: "success - empty list",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "employee_id", "name", "email", "password_hash", "role",
					"project_id", "department_id", "created_at", "updated_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			want:    []domain.User{},
			wantErr: false,
		},
		{
			name: "error - query failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users").
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

			repo := NewUserRepository(db)
			users, err := repo.FindAllUsers(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, users)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_UserRepository_DeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		setupMock func(sqlmock.Sqlmock, string)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name:   "success - delete user",
			userID: "user-1",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectExec("DELETE FROM users WHERE id = ?").
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:   "error - user not found",
			userID: "non-existent",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectExec("DELETE FROM users WHERE id = ?").
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errCode: apperr.ErrNotFound,
		},
		{
			name:   "error - foreign key constraint",
			userID: "user-with-refs",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				mock.ExpectExec("DELETE FROM users WHERE id = ?").
					WithArgs(userID).
					WillReturnError(&mysql.MySQLError{Number: 1451})
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

			tt.setupMock(mock, tt.userID)

			repo := NewUserRepository(db)
			err = repo.DeleteUser(context.Background(), tt.userID)

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
