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

func Test_mysql_ExpenseRepository_SaveExpense(t *testing.T) {
	approvedAt := time.Now()
	reviewedAt := time.Now()

	tests := []struct {
		name      string
		expense   domain.Expense
		setupMock func(sqlmock.Sqlmock, domain.Expense)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - save expense without reconciliation",
			expense: domain.Expense{
				ID:          "expense-1",
				UserID:      "user-1",
				Amount:      1000.50,
				Description: "Business trip expenses",
				Status:      domain.Pending,
				Purpose:     "Travel",
				Bills: []domain.Bill{
					{ID: "bill-1", ExpenseID: "expense-1", Amount: 500.25, Description: "Flight", AttachmentURL: "/images/flight.jpg"},
					{ID: "bill-2", ExpenseID: "expense-1", Amount: 500.25, Description: "Hotel", AttachmentURL: "/images/hotel.jpg"},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock, expense domain.Expense) {
				mock.ExpectExec("INSERT INTO expenses").
					WithArgs(expense.ID, expense.UserID, expense.Amount, expense.Description,
						expense.Status, expense.Purpose, sql.NullString{}, sql.NullTime{},
						sql.NullString{}, sql.NullTime{}, expense.IsReconciled).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect bill inserts
				for _, bill := range expense.Bills {
					mock.ExpectExec("INSERT INTO bills").
						WithArgs(bill.ID, expense.ID, bill.Amount, bill.Description, bill.AttachmentURL).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			},
			wantErr: false,
		},
		{
			name: "success - save expense with approval and reconciliation",
			expense: domain.Expense{
				ID:           "expense-2",
				UserID:       "user-1",
				Amount:       800.00,
				Description:  "Approved and reconciled expense",
				Status:       domain.Approved,
				Purpose:      "Office Supplies",
				ApprovedBy:   "manager-1",
				ApprovedAt:   approvedAt.UnixMilli(),
				ReviewedBy:   "accountant-1",
				ReviewedAt:   reviewedAt.UnixMilli(),
				IsReconciled: true,
				AdvanceID:    "advance-1",
				Bills:        []domain.Bill{},
			},
			setupMock: func(mock sqlmock.Sqlmock, expense domain.Expense) {
				mock.ExpectExec("INSERT INTO expenses").
					WithArgs(expense.ID, expense.UserID, expense.Amount, expense.Description,
						expense.Status, expense.Purpose,
						sql.NullString{String: expense.ApprovedBy, Valid: true},
						sql.NullTime{Time: time.UnixMilli(expense.ApprovedAt), Valid: true},
						sql.NullString{String: expense.ReviewedBy, Valid: true},
						sql.NullTime{Time: time.UnixMilli(expense.ReviewedAt), Valid: true},
						expense.IsReconciled).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect advance update
				mock.ExpectExec("UPDATE advances SET reconciled_expense_id = ?").
					WithArgs(expense.ID, expense.AdvanceID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error - foreign key constraint on user_id",
			expense: domain.Expense{
				ID:          "expense-3",
				UserID:      "invalid-user",
				Amount:      500.00,
				Description: "Invalid user expense",
				Status:      domain.Pending,
				Purpose:     "Test",
				Bills:       []domain.Bill{},
			},
			setupMock: func(mock sqlmock.Sqlmock, expense domain.Expense) {
				mock.ExpectExec("INSERT INTO expenses").
					WithArgs(expense.ID, expense.UserID, expense.Amount, expense.Description,
						expense.Status, expense.Purpose, sql.NullString{}, sql.NullTime{},
						sql.NullString{}, sql.NullTime{}, expense.IsReconciled).
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

			tt.setupMock(mock, tt.expense)

			repo := NewExpenseRepository(db)
			id, err := repo.SaveExpense(context.Background(), tt.expense)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expense.ID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_ExpenseRepository_UpdateExpense(t *testing.T) {
	approvedAt := time.Now()

	tests := []struct {
		name      string
		expense   domain.Expense
		setupMock func(sqlmock.Sqlmock, domain.Expense)
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name: "success - update expense",
			expense: domain.Expense{
				ID:          "expense-1",
				UserID:      "user-1",
				Amount:      1200.75,
				Description: "Updated expense",
				Status:      domain.Approved,
				Purpose:     "Travel",
				ApprovedBy:  "manager-1",
				ApprovedAt:  approvedAt.UnixMilli(),
			},
			setupMock: func(mock sqlmock.Sqlmock, expense domain.Expense) {
				mock.ExpectExec("UPDATE expenses").
					WithArgs(expense.UserID, expense.Amount, expense.Description, expense.Status,
						expense.Purpose, sql.NullString{String: expense.ApprovedBy, Valid: true},
						sql.NullTime{Time: time.UnixMilli(expense.ApprovedAt), Valid: true},
						sql.NullString{}, sql.NullTime{}, expense.IsReconciled, expense.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "success - update expense with reconciliation",
			expense: domain.Expense{
				ID:           "expense-2",
				UserID:       "user-1",
				Amount:       800.00,
				Description:  "Reconciled expense",
				Status:       domain.Approved,
				Purpose:      "Office",
				IsReconciled: true,
				AdvanceID:    "advance-1",
			},
			setupMock: func(mock sqlmock.Sqlmock, expense domain.Expense) {
				mock.ExpectExec("UPDATE expenses").
					WithArgs(expense.UserID, expense.Amount, expense.Description, expense.Status,
						expense.Purpose, sql.NullString{}, sql.NullTime{}, sql.NullString{},
						sql.NullTime{}, expense.IsReconciled, expense.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectExec("UPDATE advances SET reconciled_expense_id = ?").
					WithArgs(expense.ID, expense.AdvanceID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error - expense not found",
			expense: domain.Expense{
				ID:          "non-existent",
				UserID:      "user-1",
				Amount:      500.00,
				Description: "Ghost expense",
				Status:      domain.Pending,
				Purpose:     "Test",
			},
			setupMock: func(mock sqlmock.Sqlmock, expense domain.Expense) {
				mock.ExpectExec("UPDATE expenses").
					WithArgs(expense.UserID, expense.Amount, expense.Description, expense.Status,
						expense.Purpose, sql.NullString{}, sql.NullTime{}, sql.NullString{},
						sql.NullTime{}, expense.IsReconciled, expense.ID).
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

			tt.setupMock(mock, tt.expense)

			repo := NewExpenseRepository(db)
			err = repo.UpdateExpense(context.Background(), tt.expense)

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

func Test_mysql_ExpenseRepository_FindExpenseById(t *testing.T) {
	now := time.Now().Unix()
	approvedAt := time.Now().Unix()

	tests := []struct {
		name      string
		expenseID string
		setupMock func(sqlmock.Sqlmock, string)
		want      domain.Expense
		wantErr   bool
		errCode   apperr.AppErrorCode
	}{
		{
			name:      "success - find expense with bills",
			expenseID: "expense-1",
			setupMock: func(mock sqlmock.Sqlmock, expenseID string) {
				expenseRows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "description", "status", "purpose",
					"approved_by", "approved_at", "reviewed_by", "reviewed_at",
					"is_reconciled", "created_at", "updated_at",
				}).AddRow(
					"expense-1", "user-1", 1000.50, "Business trip", string(domain.Approved), "Travel",
					sql.NullString{String: "manager-1", Valid: true}, approvedAt,
					sql.NullString{}, int64(0), false, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM expenses WHERE id = ?").
					WithArgs(expenseID).
					WillReturnRows(expenseRows)

				billRows := sqlmock.NewRows([]string{"id", "expense_id", "amount", "description", "attachment_url"}).
					AddRow("bill-1", "expense-1", 500.25, "Flight", "/images/flight.jpg").
					AddRow("bill-2", "expense-1", 500.25, "Hotel", "/images/hotel.jpg")
				mock.ExpectQuery("SELECT (.+) FROM bills WHERE expense_id = ?").
					WithArgs(expenseID).
					WillReturnRows(billRows)
			},
			want: domain.Expense{
				ID:          "expense-1",
				UserID:      "user-1",
				Amount:      1000.50,
				Description: "Business trip",
				Status:      domain.Approved,
				Purpose:     "Travel",
				ApprovedBy:  "manager-1",
				ApprovedAt:  approvedAt,
				Bills: []domain.Bill{
					{ID: "bill-1", ExpenseID: "expense-1", Amount: 500.25, Description: "Flight", AttachmentURL: "/images/flight.jpg"},
					{ID: "bill-2", ExpenseID: "expense-1", Amount: 500.25, Description: "Hotel", AttachmentURL: "/images/hotel.jpg"},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name:      "success - find reconciled expense with advance",
			expenseID: "expense-2",
			setupMock: func(mock sqlmock.Sqlmock, expenseID string) {
				expenseRows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "description", "status", "purpose",
					"approved_by", "approved_at", "reviewed_by", "reviewed_at",
					"is_reconciled", "created_at", "updated_at",
				}).AddRow(
					"expense-2", "user-1", 800.00, "Reconciled expense", string(domain.Approved), "Office",
					sql.NullString{String: "manager-1", Valid: true}, approvedAt,
					sql.NullString{}, int64(0), true, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM expenses WHERE id = ?").
					WithArgs(expenseID).
					WillReturnRows(expenseRows)

				billRows := sqlmock.NewRows([]string{"id", "expense_id", "amount", "description", "attachment_url"})
				mock.ExpectQuery("SELECT (.+) FROM bills WHERE expense_id = ?").
					WithArgs(expenseID).
					WillReturnRows(billRows)

				advanceRows := sqlmock.NewRows([]string{"id"}).AddRow(sql.NullString{String: "advance-1", Valid: true})
				mock.ExpectQuery("SELECT id FROM advances WHERE reconciled_expense_id = ?").
					WithArgs(expenseID).
					WillReturnRows(advanceRows)
			},
			want: domain.Expense{
				ID:           "expense-2",
				UserID:       "user-1",
				Amount:       800.00,
				Description:  "Reconciled expense",
				Status:       domain.Approved,
				Purpose:      "Office",
				ApprovedBy:   "manager-1",
				ApprovedAt:   approvedAt,
				IsReconciled: true,
				AdvanceID:    "advance-1",
				Bills:        []domain.Bill{},
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			wantErr: false,
		},
		{
			name:      "error - expense not found",
			expenseID: "non-existent",
			setupMock: func(mock sqlmock.Sqlmock, expenseID string) {
				mock.ExpectQuery("SELECT (.+) FROM expenses WHERE id = ?").
					WithArgs(expenseID).
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

			tt.setupMock(mock, tt.expenseID)

			repo := NewExpenseRepository(db)
			expense, err := repo.FindExpenseById(context.Background(), tt.expenseID)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperr.AppError)
				require.True(t, ok, "expected AppError")
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, expense)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_ExpenseRepository_FindAllExpenses(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name          string
		filterOptions domain.ExpensesFilterOptions
		setupMock     func(sqlmock.Sqlmock, domain.ExpensesFilterOptions)
		wantExpenses  []domain.Expense
		wantCount     int
		wantErr       bool
	}{
		{
			name: "success - find all expenses without filter",
			filterOptions: domain.ExpensesFilterOptions{
				Limit: 10,
				Page:  1,
			},
			setupMock: func(mock sqlmock.Sqlmock, filter domain.ExpensesFilterOptions) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM expenses WHERE 1=1").
					WillReturnRows(countRows)

				expenseRows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "description", "status", "purpose",
					"approved_by", "approved_at", "reviewed_by", "reviewed_at",
					"is_reconciled", "created_at", "updated_at",
				}).
					AddRow("expense-1", "user-1", 1000.50, "Expense 1", string(domain.Pending), "Travel",
						sql.NullString{}, int64(0), sql.NullString{}, int64(0), false, now, now).
					AddRow("expense-2", "user-2", 800.00, "Expense 2", string(domain.Approved), "Office",
						sql.NullString{String: "manager-1", Valid: true}, now, sql.NullString{}, int64(0), false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM expenses WHERE 1=1 ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
					WithArgs(10, 0).
					WillReturnRows(expenseRows)

				// Bills for expense-1
				billRows1 := sqlmock.NewRows([]string{"id", "expense_id", "amount", "description", "attachment_url"})
				mock.ExpectQuery("SELECT (.+) FROM bills WHERE expense_id = ?").
					WithArgs("expense-1").
					WillReturnRows(billRows1)

				// Bills for expense-2
				billRows2 := sqlmock.NewRows([]string{"id", "expense_id", "amount", "description", "attachment_url"})
				mock.ExpectQuery("SELECT (.+) FROM bills WHERE expense_id = ?").
					WithArgs("expense-2").
					WillReturnRows(billRows2)
			},
			wantExpenses: []domain.Expense{
				{ID: "expense-1", UserID: "user-1", Amount: 1000.50, Description: "Expense 1",
					Status: domain.Pending, Purpose: "Travel", Bills: []domain.Bill{}, CreatedAt: now, UpdatedAt: now},
				{ID: "expense-2", UserID: "user-2", Amount: 800.00, Description: "Expense 2",
					Status: domain.Approved, Purpose: "Office", ApprovedBy: "manager-1", ApprovedAt: now,
					Bills: []domain.Bill{}, CreatedAt: now, UpdatedAt: now},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "success - find expenses with user filter",
			filterOptions: domain.ExpensesFilterOptions{
				UserID: "user-1",
				Limit:  10,
				Page:   1,
			},
			setupMock: func(mock sqlmock.Sqlmock, filter domain.ExpensesFilterOptions) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM expenses WHERE 1=1 AND user_id = ?").
					WithArgs("user-1").
					WillReturnRows(countRows)

				expenseRows := sqlmock.NewRows([]string{
					"id", "user_id", "amount", "description", "status", "purpose",
					"approved_by", "approved_at", "reviewed_by", "reviewed_at",
					"is_reconciled", "created_at", "updated_at",
				}).AddRow("expense-1", "user-1", 1000.50, "Expense 1", string(domain.Pending), "Travel",
					sql.NullString{}, int64(0), sql.NullString{}, int64(0), false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM expenses WHERE 1=1 AND user_id = \\? ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
					WithArgs("user-1", 10, 0).
					WillReturnRows(expenseRows)

				billRows := sqlmock.NewRows([]string{"id", "expense_id", "amount", "description", "attachment_url"})
				mock.ExpectQuery("SELECT (.+) FROM bills WHERE expense_id = ?").
					WithArgs("expense-1").
					WillReturnRows(billRows)
			},
			wantExpenses: []domain.Expense{
				{ID: "expense-1", UserID: "user-1", Amount: 1000.50, Description: "Expense 1",
					Status: domain.Pending, Purpose: "Travel", Bills: []domain.Bill{}, CreatedAt: now, UpdatedAt: now},
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

			repo := NewExpenseRepository(db)
			expenses, count, err := repo.FindAllExpenses(context.Background(), tt.filterOptions)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantExpenses, expenses)
				assert.Equal(t, tt.wantCount, count)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_mysql_ExpenseRepository_GetExpenseSumByStatus(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		status    domain.RequestStatus
		setupMock func(sqlmock.Sqlmock, string, domain.RequestStatus)
		want      float64
		wantErr   bool
	}{
		{
			name:   "success - sum all expenses",
			userID: "",
			status: "",
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(5000.75)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM expenses WHERE 1=1").
					WillReturnRows(rows)
			},
			want:    5000.75,
			wantErr: false,
		},
		{
			name:   "success - sum expenses by user",
			userID: "user-1",
			status: "",
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(3000.50)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM expenses WHERE 1=1 AND user_id = ?").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want:    3000.50,
			wantErr: false,
		},
		{
			name:   "success - sum expenses by status",
			userID: "",
			status: domain.Approved,
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(2500.25)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM expenses WHERE 1=1 AND status = ?").
					WithArgs(string(status)).
					WillReturnRows(rows)
			},
			want:    2500.25,
			wantErr: false,
		},
		{
			name:   "success - sum expenses by user and status",
			userID: "user-1",
			status: domain.Approved,
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(1200.00)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM expenses WHERE 1=1 AND user_id = \\? AND status = ?").
					WithArgs(userID, string(status)).
					WillReturnRows(rows)
			},
			want:    1200.00,
			wantErr: false,
		},
		{
			name:   "success - no expenses found",
			userID: "user-999",
			status: domain.Approved,
			setupMock: func(mock sqlmock.Sqlmock, userID string, status domain.RequestStatus) {
				rows := sqlmock.NewRows([]string{"sum"}).AddRow(0.0)
				mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM expenses WHERE 1=1 AND user_id = \\? AND status = ?").
					WithArgs(userID, string(status)).
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

			tt.setupMock(mock, tt.userID, tt.status)

			repo := NewExpenseRepository(db)
			sum, err := repo.GetExpenseSumByStatus(context.Background(), tt.userID, tt.status)

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
