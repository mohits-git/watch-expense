package mysql

import (
	"context"
	"database/sql"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(ctx context.Context, user domain.User) (string, error) {
	query := `INSERT INTO users (id, employee_id, name, email, password_hash, role, project_id, department_id) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.EmployeeId,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		nullString(user.ProjectID),
		nullString(user.DepartmentID))

	if err != nil {
		return "", HandleMysqlError(err)
	}

	return user.ID, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user domain.User) error {
	query := `UPDATE users 
			  SET employee_id = ?, name = ?, email = ?, password_hash = ?, role = ?, project_id = ?, department_id = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		user.EmployeeId,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		nullString(user.ProjectID),
		nullString(user.DepartmentID),
		user.ID)

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

func (r *UserRepository) FindUserById(ctx context.Context, userId string) (domain.User, error) {
	query := `SELECT id, employee_id, name, email, password_hash, role, project_id, department_id, 
			  UNIX_TIMESTAMP(created_at), UNIX_TIMESTAMP(updated_at)
			  FROM users WHERE id = ?`

	var user domain.User
	var projectID, departmentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.EmployeeId,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&projectID,
		&departmentID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return domain.User{}, HandleMysqlError(err)
	}

	user.ProjectID = projectID.String
	user.DepartmentID = departmentID.String

	return user, nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT id, employee_id, name, email, password_hash, role, project_id, department_id, 
			  UNIX_TIMESTAMP(created_at), UNIX_TIMESTAMP(updated_at)
			  FROM users WHERE email = ?`

	var user domain.User
	var projectID, departmentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.EmployeeId,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&projectID,
		&departmentID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return domain.User{}, HandleMysqlError(err)
	}

	user.ProjectID = projectID.String
	user.DepartmentID = departmentID.String

	return user, nil
}

func (r *UserRepository) FindAllUsers(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, employee_id, name, email, password_hash, role, project_id, department_id, 
			  UNIX_TIMESTAMP(created_at), UNIX_TIMESTAMP(updated_at)
			  FROM users`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, HandleMysqlError(err)
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		var user domain.User
		var projectID, departmentID sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.EmployeeId,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Role,
			&projectID,
			&departmentID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, HandleMysqlError(err)
		}

		user.ProjectID = projectID.String
		user.DepartmentID = departmentID.String

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, HandleMysqlError(err)
	}

	return users, nil
}
