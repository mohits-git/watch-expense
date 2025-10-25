package mysql

import (
	"context"
	"database/sql"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) ports.ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) SaveExpense(ctx context.Context, expense domain.Expense) (string, error) {
	query := `INSERT INTO expenses (id, user_id, amount, description, status, purpose, approved_by, approved_at, reviewed_by, reviewed_at, is_reconciled) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		expense.ID,
		expense.UserID,
		expense.Amount,
		expense.Description,
		expense.Status,
		expense.Purpose,
		nullString(expense.ApprovedBy),
		nullInt64(expense.ApprovedAt),
		nullString(expense.ReviewedBy),
		nullInt64(expense.ReviewedAt),
		expense.IsReconciled)

	if err != nil {
		return "", HandleMysqlError(err)
	}

	return expense.ID, nil
}

func (r *ExpenseRepository) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	query := `UPDATE expenses 
			  SET user_id = ?, amount = ?, description = ?, status = ?, purpose = ?, 
			      approved_by = ?, approved_at = ?, reviewed_by = ?, reviewed_at = ?, is_reconciled = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		expense.UserID,
		expense.Amount,
		expense.Description,
		expense.Status,
		expense.Purpose,
		nullString(expense.ApprovedBy),
		nullInt64(expense.ApprovedAt),
		nullString(expense.ReviewedBy),
		nullInt64(expense.ReviewedAt),
		expense.IsReconciled,
		expense.ID)

	if err != nil {
		return HandleMysqlError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return HandleMysqlError(err)
	}

	if rowsAffected == 0 {
		return HandleMysqlError(sql.ErrNoRows)
	}

	return nil
}

func (r *ExpenseRepository) FindExpenseById(ctx context.Context, expenseId string) (domain.Expense, error) {
	query := `SELECT id, user_id, amount, description, status, purpose, 
			  approved_by, 
			  IFNULL(UNIX_TIMESTAMP(approved_at), 0), 
			  reviewed_by, 
			  IFNULL(UNIX_TIMESTAMP(reviewed_at), 0),
			  is_reconciled,
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM expenses WHERE id = ?`

	var expense domain.Expense
	var approvedBy, reviewedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, expenseId).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Description,
		&expense.Status,
		&expense.Purpose,
		&approvedBy,
		&expense.ApprovedAt,
		&reviewedBy,
		&expense.ReviewedAt,
		&expense.IsReconciled,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	if err != nil {
		return domain.Expense{}, HandleMysqlError(err)
	}

	expense.ApprovedBy = approvedBy.String
	expense.ReviewedBy = reviewedBy.String

	bills, err := r.findBillsByExpenseId(ctx, expenseId)
	if err != nil {
		return domain.Expense{}, err
	}
	expense.Bills = bills

	if expense.IsReconciled {
		advanceID, err := r.findAdvanceIdByExpenseId(ctx, expenseId)
		if err != nil {
			return domain.Expense{}, err
		}
		expense.AdvanceID = advanceID
	}

	return expense, nil
}

func (r *ExpenseRepository) FindAllExpenses(ctx context.Context, filterOptions domain.ExpensesFilterOptions) ([]domain.Expense, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}

	if filterOptions.UserID != "" {
		whereClause += ` AND user_id = ?`
		args = append(args, filterOptions.UserID)
	}

	if filterOptions.Status != "" {
		whereClause += ` AND status = ?`
		args = append(args, filterOptions.Status)
	}

	countQuery := `SELECT COUNT(*) FROM expenses ` + whereClause
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, HandleMysqlError(err)
	}

	query := `SELECT id, user_id, amount, description, status, purpose, 
			  approved_by, 
			  IFNULL(UNIX_TIMESTAMP(approved_at), 0), 
			  reviewed_by, 
			  IFNULL(UNIX_TIMESTAMP(reviewed_at), 0),
			  is_reconciled,
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM expenses ` + whereClause

	query += ` ORDER BY created_at DESC`

	if filterOptions.Limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		offset := 0
		if filterOptions.Page > 0 {
			offset = (filterOptions.Page - 1) * filterOptions.Limit
		}
		args = append(args, filterOptions.Limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, HandleMysqlError(err)
	}
	defer rows.Close()

	expenses := []domain.Expense{}
	for rows.Next() {
		var expense domain.Expense
		var approvedBy, reviewedBy sql.NullString

		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.Amount,
			&expense.Description,
			&expense.Status,
			&expense.Purpose,
			&approvedBy,
			&expense.ApprovedAt,
			&reviewedBy,
			&expense.ReviewedAt,
			&expense.IsReconciled,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)

		if err != nil {
			return nil, 0, HandleMysqlError(err)
		}

		expense.ApprovedBy = approvedBy.String
		expense.ReviewedBy = reviewedBy.String

		bills, err := r.findBillsByExpenseId(ctx, expense.ID)
		if err != nil {
			return nil, 0, err
		}
		expense.Bills = bills

		if expense.IsReconciled {
			advanceID, err := r.findAdvanceIdByExpenseId(ctx, expense.ID)
			if err != nil {
				return nil, 0, err
			}
			expense.AdvanceID = advanceID
		}

		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, HandleMysqlError(err)
	}

	return expenses, totalCount, nil
}

func (r *ExpenseRepository) findBillsByExpenseId(ctx context.Context, expenseId string) ([]domain.Bill, error) {
	query := `SELECT id, expense_id, amount, description, attachment_url
			  FROM bills WHERE expense_id = ?`

	rows, err := r.db.QueryContext(ctx, query, expenseId)
	if err != nil {
		return nil, HandleMysqlError(err)
	}
	defer rows.Close()

	bills := []domain.Bill{}
	for rows.Next() {
		var bill domain.Bill

		err := rows.Scan(
			&bill.ID,
			&bill.ExpenseID,
			&bill.Amount,
			&bill.Description,
			&bill.AttachmentURL,
		)

		if err != nil {
			return nil, HandleMysqlError(err)
		}

		bills = append(bills, bill)
	}

	if err = rows.Err(); err != nil {
		return nil, HandleMysqlError(err)
	}

	return bills, nil
}

func (r *ExpenseRepository) findAdvanceIdByExpenseId(ctx context.Context, expenseId string) (string, error) {
	query := `SELECT id FROM advances WHERE reconciled_expense_id = ? LIMIT 1`

	var advanceID sql.NullString
	err := r.db.QueryRowContext(ctx, query, expenseId).Scan(&advanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", HandleMysqlError(err)
	}

	return advanceID.String, nil
}
