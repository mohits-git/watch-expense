package dtos

type Project struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Budget           float64 `json:"budget"`
	StartDate        int64   `json:"startDate"`
	EndDate          int64   `json:"endDate"`
	ProjectManagerID string  `json:"projectManagerId"`
	DepartmentID     string  `json:"departmentId"`
	CreatedAt        int64   `json:"createdAt"`
	UpdatedAt        int64   `json:"updatedAt"`
}
