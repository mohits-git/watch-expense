package mysql

import (
	"context"
	"database/sql"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type DepartmentRepository struct {
	db *sql.DB
}

func NewDepartmentRepository(db *sql.DB) ports.DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) SaveDepartment(ctx context.Context, department domain.Department) (string, error) {
	query := `INSERT INTO departments (id, name, budget) 
			  VALUES (?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		department.ID,
		department.Name,
		department.Budget)

	if err != nil {
		return "", HandleMysqlError(err)
	}

	return department.ID, nil
}

func (r *DepartmentRepository) UpdateDepartment(ctx context.Context, department domain.Department) error {
	query := `UPDATE departments 
			  SET name = ?, budget = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		department.Name,
		department.Budget,
		department.ID)

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

func (r *DepartmentRepository) FindDepartmentById(ctx context.Context, departmentId string) (domain.Department, error) {
	query := `SELECT id, name, budget, UNIX_TIMESTAMP(created_at), UNIX_TIMESTAMP(updated_at)
			  FROM departments WHERE id = ?`

	var department domain.Department

	err := r.db.QueryRowContext(ctx, query, departmentId).Scan(
		&department.ID,
		&department.Name,
		&department.Budget,
		&department.CreatedAt,
		&department.UpdatedAt,
	)

	if err != nil {
		return domain.Department{}, HandleMysqlError(err)
	}

	return department, nil
}

func (r *DepartmentRepository) FindAllDepartments(ctx context.Context) ([]domain.Department, error) {
	query := `SELECT id, name, budget, UNIX_TIMESTAMP(created_at), UNIX_TIMESTAMP(updated_at)
			  FROM departments`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, HandleMysqlError(err)
	}
	defer rows.Close()

	departments := []domain.Department{}
	for rows.Next() {
		var department domain.Department

		err := rows.Scan(
			&department.ID,
			&department.Name,
			&department.Budget,
			&department.CreatedAt,
			&department.UpdatedAt,
		)

		if err != nil {
			return nil, HandleMysqlError(err)
		}

		departments = append(departments, department)
	}

	if err = rows.Err(); err != nil {
		return nil, HandleMysqlError(err)
	}

	return departments, nil
}
