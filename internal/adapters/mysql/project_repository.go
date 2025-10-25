package mysql

import (
	"context"
	"database/sql"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ports.ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) SaveProject(ctx context.Context, project domain.Project) (string, error) {
	query := `INSERT INTO projects (id, name, description, budget, start_date, end_date, department_id) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		project.ID,
		project.Name,
		project.Description,
		project.Budget,
		nullInt64(project.StartDate),
		nullInt64(project.EndDate),
		nullString(project.DepartmentID))

	if err != nil {
		return "", HandleMysqlError(err)
	}

	return project.ID, nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, project domain.Project) error {
	query := `UPDATE projects 
			  SET name = ?, description = ?, budget = ?, start_date = ?, end_date = ?, department_id = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		project.Name,
		project.Description,
		project.Budget,
		nullInt64(project.StartDate),
		nullInt64(project.EndDate),
		nullString(project.DepartmentID),
		project.ID)

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

func (r *ProjectRepository) FindProjectById(ctx context.Context, projectId string) (domain.Project, error) {
	query := `SELECT id, name, description, budget, 
			  IFNULL(UNIX_TIMESTAMP(start_date), 0), 
			  IFNULL(UNIX_TIMESTAMP(end_date), 0), 
			  department_id, 
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM projects WHERE id = ?`

	var project domain.Project
	var departmentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, projectId).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Budget,
		&project.StartDate,
		&project.EndDate,
		&departmentID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return domain.Project{}, HandleMysqlError(err)
	}

	project.DepartmentID = departmentID.String

	return project, nil
}

func (r *ProjectRepository) FindAllProjects(ctx context.Context) ([]domain.Project, error) {
	query := `SELECT id, name, description, budget, 
			  IFNULL(UNIX_TIMESTAMP(start_date), 0), 
			  IFNULL(UNIX_TIMESTAMP(end_date), 0), 
			  department_id, 
			  UNIX_TIMESTAMP(created_at), 
			  UNIX_TIMESTAMP(updated_at)
			  FROM projects`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, HandleMysqlError(err)
	}
	defer rows.Close()

	projects := []domain.Project{}
	for rows.Next() {
		var project domain.Project
		var departmentID sql.NullString

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Budget,
			&project.StartDate,
			&project.EndDate,
			&departmentID,
			&project.CreatedAt,
			&project.UpdatedAt,
		)

		if err != nil {
			return nil, HandleMysqlError(err)
		}

		project.DepartmentID = departmentID.String

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, HandleMysqlError(err)
	}

	return projects, nil
}
