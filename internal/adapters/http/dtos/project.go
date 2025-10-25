package dtos

import "github.com/mohits-git/watch-expense/internal/domain"

type Project struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Budget       float64 `json:"budget"`
	StartDate    int64   `json:"startDate"`
	EndDate      int64   `json:"endDate"`
	DepartmentID string  `json:"departmentId"`
	CreatedAt    int64   `json:"createdAt"`
	UpdatedAt    int64   `json:"updatedAt"`
}

func ToProjectDTO(projectDomain domain.Project) Project {
	return Project{
		ID:           projectDomain.ID,
		Name:         projectDomain.Name,
		Description:  projectDomain.Description,
		Budget:       projectDomain.Budget,
		StartDate:    projectDomain.StartDate,
		EndDate:      projectDomain.EndDate,
		DepartmentID: projectDomain.DepartmentID,
		CreatedAt:    projectDomain.CreatedAt,
		UpdatedAt:    projectDomain.UpdatedAt,
	}
}

func ToProjectDomain(projectDTO Project) domain.Project {
	return domain.Project{
		ID:           projectDTO.ID,
		Name:         projectDTO.Name,
		Description:  projectDTO.Description,
		Budget:       projectDTO.Budget,
		StartDate:    projectDTO.StartDate,
		EndDate:      projectDTO.EndDate,
		DepartmentID: projectDTO.DepartmentID,
		CreatedAt:    projectDTO.CreatedAt,
		UpdatedAt:    projectDTO.UpdatedAt,
	}
}

type CreateProjectRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Budget       float64 `json:"budget"`
	StartDate    int64   `json:"startDate"`
	EndDate      int64   `json:"endDate"`
	DepartmentID string  `json:"departmentId"`
}

type UpdateProjectRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Budget       float64 `json:"budget"`
	StartDate    int64   `json:"startDate"`
	EndDate      int64   `json:"endDate"`
	DepartmentID string  `json:"departmentId"`
}

type CreateProjectResponse struct {
	ID string `json:"id"`
}
