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

func Test_mysql_AdvanceRepository_SaveAdvance(t *testing.T) {
	approvedAt := time.Now()
	reviewedAt := time.Now()

	tests := []struct {
		name      string
		advance   domain.Advance
		setupMock func(sqlmock.Sqlmock, domain.Advance)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - save advance without approval",
			advance: domain.Advance{
				ID:          "advance-1",
				UserID:      "user-1",
				Amount:      5000.00,
				Purpose:     "Business trip advance",
				Description: "Advance for upcoming trip",
				Status:      domain.Pending,
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("INSERT INTO advances").
					WithArgs(advance.ID, advance.UserID, advance.Amount, advance.Purpose,
						advance.Description, advance.Status, sql.NullString{},
						sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullTime{}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "success - save advance with approval",
			advance: domain.Advance{
				ID:          "advance-2",
				UserID:      "user-1",
				Amount:      3000.00,
				Purpose:     "Office supplies advance",
				Description: "Advance for office equipment",
				Status:      domain.Approved,
				ApprovedBy:  "manager-1",
				ApprovedAt:  approvedAt.UnixMilli(),
				ReviewedBy:  "accountant-1",
				ReviewedAt:  reviewedAt.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("INSERT INTO advances").
					WithArgs(advance.ID, advance.UserID, advance.Amount, advance.Purpose,
						advance.Description, advance.Status, sql.NullString{},
						sql.NullString{String: advance.ApprovedBy, Valid: true},
						sql.NullTime{Time: time.UnixMilli(advance.ApprovedAt), Valid: true},
						sql.NullString{String: advance.ReviewedBy, Valid: true},
						sql.NullTime{Time: time.UnixMilli(advance.ReviewedAt), Valid: true}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "success - save advance with reconciled expense",
			advance: domain.Advance{
				ID:                  "advance-3",
				UserID:              "user-1",
				Amount:              2000.00,
				Purpose:             "Reconciled advance",
				Description:         "Already reconciled",
				Status:              domain.Approved,
				ReconciledExpenseID: "expense-1",
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("INSERT INTO advances").
					WithArgs(advance.ID, advance.UserID, advance.Amount, advance.Purpose,
						advance.Description, advance.Status,
						sql.NullString{String: advance.ReconciledExpenseID, Valid: true},
						sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullTime{}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "error - foreign key constraint on user_id",
			advance: domain.Advance{
				ID:          "advance-4",
				UserID:      "invalid-user",
				Amount:      1000.00,
				Purpose:     "Invalid user advance",
				Description: "Should fail",
				Status:      domain.Pending,
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("INSERT INTO advances").
					WithArgs(advance.ID, advance.UserID, advance.Amount, advance.Purpose,
						advance.Description, advance.Status, sql.NullString{},
						sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullTime{}).
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

			tt.setupMock(mock, tt.advance)

			repo := NewAdvanceRepository(db)
			id, err := repo.SaveAdvance(context.Background(), tt.advance)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.advance.ID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_AdvanceRepository_UpdateAdvance(t *testing.T) {
	approvedAt := time.Now()

	tests := []struct {
		name      string
		advance   domain.Advance
		setupMock func(sqlmock.Sqlmock, domain.Advance)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - update advance",
			advance: domain.Advance{
				ID:          "advance-1",
				UserID:      "user-1",
				Amount:      6000.00,
				Purpose:     "Updated advance",
				Description: "Updated description",
				Status:      domain.Approved,
				ApprovedBy:  "manager-1",
				ApprovedAt:  approvedAt.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("UPDATE advances").
					WithArgs(advance.UserID, advance.Amount, advance.Purpose, advance.Description,
						advance.Status, sql.NullString{},
						sql.NullString{String: advance.ApprovedBy, Valid: true},
						sql.NullTime{Time: time.UnixMilli(advance.ApprovedAt), Valid: true},
						sql.NullString{}, sql.NullTime{}, advance.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "success - update advance with reconciliation",
			advance: domain.Advance{
				ID:                  "advance-2",
				UserID:              "user-1",
				Amount:              4000.00,
				Purpose:             "Reconciled advance",
				Description:         "Now reconciled",
				Status:              domain.Approved,
				ReconciledExpenseID: "expense-1",
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("UPDATE advances").
					WithArgs(advance.UserID, advance.Amount, advance.Purpose, advance.Description,
						advance.Status, sql.NullString{String: advance.ReconciledExpenseID, Valid: true},
						sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullTime{}, advance.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error - advance not found",
			advance: domain.Advance{
				ID:          "non-existent",
				UserID:      "user-1",
				Amount:      1000.00,
				Purpose:     "Ghost advance",
				Description: "Does not exist",
				Status:      domain.Pending,
			},
			setupMock: func(mock sqlmock.Sqlmock, advance domain.Advance) {
				mock.ExpectExec("UPDATE advances").
					WithArgs(advance.UserID, advance.Amount, advance.Purpose, advance.Description,
						advance.Status, sql.NullString{}, sql.NullString{}, sql.NullTime{},
						sql.NullString{}, sql.NullTime{}, advance.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
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

			tt.setupMock(mock, tt.advance)

			repo := NewAdvanceRepository(db)
			err = repo.UpdateAdvance(context.Background(), tt.advance)

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

func Test_mysql_AdvanceRepository_FindAdvanceById(t *testing.T) {
	now := time.Now().Unix()
	approvedAt := time.Now().Unix()

	tests := []struct {
		name      string
		advanceID string
		setupMock func(sqlmock.Sqlmock, string)
		want      domain.Advance
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name:      "success - find advance without reconciliation",
			advanceID: "advance-1",
			setupMock: func(mock sqlmock.Sqlmock, advanceID string) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "purpose", "description", "status",
					"reconciled_expense_id", "approved_by", "approved_at", "reviewed_by",
					"reviewed_at", "created_at", "updated_at",
				}).AddRow(
					"advance-1", "user-1", 5000.00, "Business trip", "Trip advance", string(domain.Approved),
					sql.NullString{}, sql.NullString{String: "manager-1", Valid: true}, approvedAt,
					sql.NullString{}, int64(0), now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM advances WHERE id = ?").
					WithArgs(advanceID).
					WillReturnRows(rows)
			},
			want: domain.Advance{
				ID:          "advance-1",
				UserID:      "user-1",
				Amount:      5000.00,
				Purpose:     "Business trip",
				Description: "Trip advance",
				Status:      domain.Approved,
				ApprovedBy:  "manager-1",
				ApprovedAt:  approvedAt,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: false,
		},
		{
			name:      "success - find advance with reconciliation",
			advanceID: "advance-2",
			setupMock: func(mock sqlmock.Sqlmock, advanceID string) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "purpose", "description", "status",
					"reconciled_expense_id", "approved_by", "approved_at", "reviewed_by",
					"reviewed_at", "created_at", "updated_at",
				}).AddRow(
					"advance-2", "user-1", 3000.00, "Office supplies", "Advance for supplies", string(domain.Approved),
					sql.NullString{String: "expense-1", Valid: true},
					sql.NullString{String: "manager-1", Valid: true}, approvedAt,
					sql.NullString{String: "accountant-1", Valid: true}, approvedAt, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM advances WHERE id = ?").
					WithArgs(advanceID).
					WillReturnRows(rows)
			},
			want: domain.Advance{
				ID:                  "advance-2",
				UserID:              "user-1",
				Amount:              3000.00,
				Purpose:             "Office supplies",
				Description:         "Advance for supplies",
				Status:              domain.Approved,
				ReconciledExpenseID: "expense-1",
				ApprovedBy:          "manager-1",
				ApprovedAt:          approvedAt,
				ReviewedBy:          "accountant-1",
				ReviewedAt:          approvedAt,
				CreatedAt:           now,
				UpdatedAt:           now,
			},
			wantErr: false,
		},
		{
			name:      "error - advance not found",
			advanceID: "non-existent",
			setupMock: func(mock sqlmock.Sqlmock, advanceID string) {
				mock.ExpectQuery("SELECT (.+) FROM advances WHERE id = ?").
					WithArgs(advanceID).
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

			tt.setupMock(mock, tt.advanceID)

			repo := NewAdvanceRepository(db)
			advance, err := repo.FindAdvanceById(context.Background(), tt.advanceID)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, advance)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_AdvanceRepository_FindAllAdvances(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name          string
		filterOptions domain.AdvancesFilterOptions
		setupMock     func(sqlmock.Sqlmock, domain.AdvancesFilterOptions)
		wantAdvances  []domain.Advance
		wantCount     int
		wantErr       bool
	}{
		{
			name: "success - find all advances without filter",
			filterOptions: domain.AdvancesFilterOptions{
				Limit: 10,
				Page:  1,
			},
			setupMock: func(mock sqlmock.Sqlmock, filter domain.AdvancesFilterOptions) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM advances WHERE 1=1").
					WillReturnRows(countRows)

				advanceRows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "purpose", "description", "status",
					"reconciled_expense_id", "approved_by", "approved_at", "reviewed_by",
					"reviewed_at", "created_at", "updated_at",
				}).
					AddRow("advance-1", "user-1", 5000.00, "Trip", "Business trip", string(domain.Pending),
						sql.NullString{}, sql.NullString{}, int64(0), sql.NullString{}, int64(0), now, now).
					AddRow("advance-2", "user-2", 3000.00, "Supplies", "Office supplies", string(domain.Approved),
						sql.NullString{}, sql.NullString{String: "manager-1", Valid: true}, now,
						sql.NullString{}, int64(0), now, now)

				mock.ExpectQuery("SELECT (.+) FROM advances WHERE 1=1 ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
					WithArgs(10, 0).
					WillReturnRows(advanceRows)
			},
			wantAdvances: []domain.Advance{
				{ID: "advance-1", UserID: "user-1", Amount: 5000.00, Purpose: "Trip",
					Description: "Business trip", Status: domain.Pending, CreatedAt: now, UpdatedAt: now},
				{ID: "advance-2", UserID: "user-2", Amount: 3000.00, Purpose: "Supplies",
					Description: "Office supplies", Status: domain.Approved, ApprovedBy: "manager-1",
					ApprovedAt: now, CreatedAt: now, UpdatedAt: now},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "success - find advances with user filter",
			filterOptions: domain.AdvancesFilterOptions{
				UserID: "user-1",
				Limit:  10,
				Page:   1,
			},
			setupMock: func(mock sqlmock.Sqlmock, filter domain.AdvancesFilterOptions) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM advances WHERE 1=1 AND user_id = ?").
					WithArgs("user-1").
					WillReturnRows(countRows)

				advanceRows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "purpose", "description", "status",
					"reconciled_expense_id", "approved_by", "approved_at", "reviewed_by",
					"reviewed_at", "created_at", "updated_at",
				}).AddRow("advance-1", "user-1", 5000.00, "Trip", "Business trip", string(domain.Pending),
					sql.NullString{}, sql.NullString{}, int64(0), sql.NullString{}, int64(0), now, now)

				mock.ExpectQuery("SELECT (.+) FROM advances WHERE 1=1 AND user_id = \\? ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
					WithArgs("user-1", 10, 0).
					WillReturnRows(advanceRows)
			},
			wantAdvances: []domain.Advance{
				{ID: "advance-1", UserID: "user-1", Amount: 5000.00, Purpose: "Trip",
					Description: "Business trip", Status: domain.Pending, CreatedAt: now, UpdatedAt: now},
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock, tt.filterOptions)

			repo := NewAdvanceRepository(db)
			advances, count, err := repo.FindAllAdvances(context.Background(), tt.filterOptions)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantAdvances, advances)
				assert.Equal(t, tt.wantCount, count)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_AdvanceRepository_GetAdvanceSumByStatus(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		status    domain.RequestStatus
		setupMock func(sqlmock.Sqlmock, string, domain.RequestStatus)
		want      float64
		wantErr   bool
	}{
		{
			name:   "success - sum all advances",
			userID: "",
			status: "",
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(8000.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE 1=1").
					WillReturnRows(rows)
			},
			want:    8000.00,
			wantErr: false,
		},
		{
			name:   "success - sum advances by user",
			userID: "user-1",
			status: "",
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(5000.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE 1=1 AND user_id = ?").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    5000.00,
			wantErr: false,
		},
		{
			name:   "success - sum advances by status",
			userID: "",
			status: domain.Approved,
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(3000.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE 1=1 AND status = ?").
					WithArgs(string(status)).
					WillReturnRows(rows)
			},
			want:    3000.00,
			wantErr: false,
		},
		{
			name:   "success - sum advances by user and status",
			userID: "user-1",
			status: domain.Approved,
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(2500.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE 1=1 AND user_id = \\? AND status = ?").
					WithArgs(userID, string(status)).
					WillReturnRows(rows)
			},
			want:    2500.00,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock, tt.userID, tt.status)

			repo := NewAdvanceRepository(db)
			sum, err := repo.GetAdvanceSumByStatus(context.Background(), tt.userID, tt.status)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, sum)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_AdvanceRepository_GetReconciledAdvancesSum(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		setupMock func(sqlmock.Sqlmock, string)
		want      float64
		wantErr   bool
	}{
		{
			name:   "success - sum all reconciled advances",
			userID: "",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(4500.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE reconciled_expense_id IS NOT NULL AND reconciled_expense_id != ''").
					WillReturnRows(rows)
			},
			want:    4500.00,
			wantErr: false,
		},
		{
			name:   "success - sum reconciled advances by user",
			userID: "user-1",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(3000.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE reconciled_expense_id IS NOT NULL AND reconciled_expense_id != '' AND user_id = ?").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    3000.00,
			wantErr: false,
		},
		{
			name:   "success - no reconciled advances",
			userID: "user-999",
			setupMock: func(mock sqlmock.Sqlmock, userID string) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(0.0)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM advances WHERE reconciled_expense_id IS NOT NULL AND reconciled_expense_id != '' AND user_id = ?").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    0.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock, tt.userID)

			repo := NewAdvanceRepository(db)
			sum, err := repo.GetReconciledAdvancesSum(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, sum)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
