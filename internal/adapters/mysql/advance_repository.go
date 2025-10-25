package mysql

import (
	"context"
	"database/sql"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type AdvanceRepository struct {
	db *sql.DB
}

func NewAdvanceRepository(db *sql.DB) ports.AdvanceRepository {
	return &AdvanceRepository{db: db}
}

func (r *AdvanceRepository) SaveAdvance(ctx context.Context, advance domain.Advance) (string, error) {
	query := `INSERT INTO advances (id, user_id, amount, purpose, description, status, reconciled_expense_id, approved_by, approved_at, reviewed_by, reviewed_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		advance.ID,
		advance.UserID,
		advance.Amount,
		advance.Purpose,
		advance.Description,
		advance.Status,
		nullString(advance.ReconciledExpenseID),
		nullString(advance.ApprovedBy),
		nullInt64(advance.ApprovedAt),
		nullString(advance.ReviewedBy),
		nullInt64(advance.ReviewedAt))

	if err != nil {
		return "", HandleMysqlError(err)
	}

	return advance.ID, nil
}

func (r *AdvanceRepository) UpdateAdvance(ctx context.Context, advance domain.Advance) error {
	query := `UPDATE advances 
			  SET user_id = ?, amount = ?, purpose = ?, description = ?, status = ?, 
			      reconciled_expense_id = ?, approved_by = ?, approved_at = ?, reviewed_by = ?, reviewed_at = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		advance.UserID,
		advance.Amount,
		advance.Purpose,
		advance.Description,
		advance.Status,
		nullString(advance.ReconciledExpenseID),
		nullString(advance.ApprovedBy),
		nullInt64(advance.ApprovedAt),
		nullString(advance.ReviewedBy),
		nullInt64(advance.ReviewedAt),
		advance.ID)

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

func (r *AdvanceRepository) FindAdvanceById(ctx context.Context, advanceId string) (domain.Advance, error) {
	query := `SELECT id, user_id, amount, purpose, description, status, 
			  reconciled_expense_id, 
			  approved_by, 
			  IFNULL(UNIX_TIMESTAMP(approved_at), 0), 
			  reviewed_by, 
			  IFNULL(UNIX_TIMESTAMP(reviewed_at), 0),
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM advances WHERE id = ?`

	var advance domain.Advance
	var reconciledExpenseID, approvedBy, reviewedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, advanceId).Scan(
		&advance.ID,
		&advance.UserID,
		&advance.Amount,
		&advance.Purpose,
		&advance.Description,
		&advance.Status,
		&reconciledExpenseID,
		&approvedBy,
		&advance.ApprovedAt,
		&reviewedBy,
		&advance.ReviewedAt,
		&advance.CreatedAt,
		&advance.UpdatedAt,
	)

	if err != nil {
		return domain.Advance{}, HandleMysqlError(err)
	}

	advance.ReconciledExpenseID = reconciledExpenseID.String
	advance.ApprovedBy = approvedBy.String
	advance.ReviewedBy = reviewedBy.String

	return advance, nil
}

func (r *AdvanceRepository) FindAdvancesByUserId(ctx context.Context, userId string) ([]domain.Advance, error) {
	query := `SELECT id, user_id, amount, purpose, description, status, 
			  reconciled_expense_id, 
			  approved_by, 
			  IFNULL(UNIX_TIMESTAMP(approved_at), 0), 
			  reviewed_by, 
			  IFNULL(UNIX_TIMESTAMP(reviewed_at), 0),
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM advances WHERE user_id = ?`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, HandleMysqlError(err)
	}
	defer rows.Close()

	advances := []domain.Advance{}
	for rows.Next() {
		var advance domain.Advance
		var reconciledExpenseID, approvedBy, reviewedBy sql.NullString

		err := rows.Scan(
			&advance.ID,
			&advance.UserID,
			&advance.Amount,
			&advance.Purpose,
			&advance.Description,
			&advance.Status,
			&reconciledExpenseID,
			&approvedBy,
			&advance.ApprovedAt,
			&reviewedBy,
			&advance.ReviewedAt,
			&advance.CreatedAt,
			&advance.UpdatedAt,
		)

		if err != nil {
			return nil, HandleMysqlError(err)
		}

		advance.ReconciledExpenseID = reconciledExpenseID.String
		advance.ApprovedBy = approvedBy.String
		advance.ReviewedBy = reviewedBy.String

		advances = append(advances, advance)
	}

	if err = rows.Err(); err != nil {
		return nil, HandleMysqlError(err)
	}

	return advances, nil
}

func (r *AdvanceRepository) FindAllAdvances(ctx context.Context, filterOptions domain.AdvancesFilterOptions) ([]domain.Advance, int, error) {
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

	countQuery := `SELECT COUNT(*) FROM advances ` + whereClause
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, HandleMysqlError(err)
	}

	query := `SELECT id, user_id, amount, purpose, description, status, 
			  reconciled_expense_id, 
			  approved_by, 
			  IFNULL(UNIX_TIMESTAMP(approved_at), 0), 
			  reviewed_by, 
			  IFNULL(UNIX_TIMESTAMP(reviewed_at), 0),
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM advances ` + whereClause

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

	advances := []domain.Advance{}
	for rows.Next() {
		var advance domain.Advance
		var reconciledExpenseID, approvedBy, reviewedBy sql.NullString

		err := rows.Scan(
			&advance.ID,
			&advance.UserID,
			&advance.Amount,
			&advance.Purpose,
			&advance.Description,
			&advance.Status,
			&reconciledExpenseID,
			&approvedBy,
			&advance.ApprovedAt,
			&reviewedBy,
			&advance.ReviewedAt,
			&advance.CreatedAt,
			&advance.UpdatedAt,
		)

		if err != nil {
			return nil, 0, HandleMysqlError(err)
		}

		advance.ReconciledExpenseID = reconciledExpenseID.String
		advance.ApprovedBy = approvedBy.String
		advance.ReviewedBy = reviewedBy.String

		advances = append(advances, advance)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, HandleMysqlError(err)
	}

	return advances, totalCount, nil
}
