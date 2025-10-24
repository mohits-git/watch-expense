package bcrypt

import (
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	} else if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}
	return &BcryptPasswordHasher{
		cost: cost,
	}
}

func (h *BcryptPasswordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		if err == bcrypt.ErrPasswordTooLong {
			return "", apperr.NewAppError(apperr.ErrInvalid, "password too long", err)
		}
		return "", err
	}
	return string(hash), nil
}

func (h *BcryptPasswordHasher) ComparePassword(hashedPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return false, nil
		case bcrypt.ErrHashTooShort:
			return false, apperr.NewAppError(apperr.ErrInvalid, "hashed password too short", err)
		default:
			return false, err
		}
	}
	return true, nil
}
