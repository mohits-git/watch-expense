package dtos

import "github.com/mohits-git/watch-expense/internal/domain"

type Department struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Budget    float64 `json:"budget"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
}

func ToDepartmentDTO(departmentDomain domain.Department) Department {
  return Department{
    ID:        departmentDomain.ID,
    Name:      departmentDomain.Name,
    Budget:    departmentDomain.Budget,
    CreatedAt: departmentDomain.CreatedAt,
    UpdatedAt: departmentDomain.UpdatedAt,
  }
}

func ToDepartmentDomain(departmentDTO Department) domain.Department {
  return domain.Department{
    ID:        departmentDTO.ID,
    Name:      departmentDTO.Name,
    Budget:    departmentDTO.Budget,
    CreatedAt: departmentDTO.CreatedAt,
    UpdatedAt: departmentDTO.UpdatedAt,
  }
}

type CreateDepartmentRequest struct {
  Name   string  `json:"name"`
  Budget float64 `json:"budget"`
}

type UpdateDepartmentRequest struct {
  Name   string  `json:"name"`
  Budget float64 `json:"budget"`
}

type CreateDepartmentResponse struct {
  ID        string  `json:"id"`
}
