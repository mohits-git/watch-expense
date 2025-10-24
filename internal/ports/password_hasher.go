package ports

type PasswordHasher interface {
	HashPassword(password string) (hashedPassword string, err error)
	ComparePassword(hashedPassword, password string) (isValid bool, err error)
}
