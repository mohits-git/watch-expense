package mockpasswordhasher

import "github.com/stretchr/testify/mock"

type PasswordHasher struct {
	mock.Mock
}

func (p *PasswordHasher) HashPassword(password string) (hashedPassword string, err error) {
	args := p.Called(password)
	return args.String(0), args.Error(1)
}

func (p *PasswordHasher) ComparePassword(hashedPassword, password string) (isValid bool, err error) {
	args := p.Called(hashedPassword, password)
	return args.Bool(0), args.Error(1)
}
