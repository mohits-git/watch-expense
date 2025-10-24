package domain

type Project struct {
	ID               string
	Name             string
	Description      string
	Budget           float64
	StartDate        int64
	EndDate          int64
	ProjectManagerID string
	DepartmentID     string
	CreatedAt        int64
	UpdatedAt        int64
}
