package domain

type Department struct {
	ID        string
	Name      string
	Budget    float64
	ManagerID string
	CreatedAt int64
	UpdatedAt int64
}
