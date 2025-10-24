package domain

type UserRole string

const (
	Admin    UserRole = "ADMIN"
	Employee UserRole = "EMPLOYEE"
	Manager  UserRole = "MANAGER"
)

type User struct {
	ID           string
	EmployeeId   string
	Name         string
	Password     string
	Email        string
	Role         UserRole
	ProjectID    string
	DepartmentID string
	CreatedAt    int64
	UpdatedAt    int64
}
