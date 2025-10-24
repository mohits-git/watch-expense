package dtos

type Department struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Budget    float64 `json:"budget"`
	ManagerID string  `json:"managerId"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
}
