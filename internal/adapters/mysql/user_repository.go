package mysql

import (
	"context"
	"database/sql"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) SaveUser(ctx context.Context, user domain.User) (string, error) {
  // Implementation goes here
  return "", nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user domain.User) error {
  // Implementation goes here
  return nil
}

func (r *UserRepository) FindUserById(ctx context.Context, userId string) (domain.User, error) {
  // Implementation goes here
  return domain.User{}, nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
  // Implementation goes here
  return domain.User{}, nil
}

func (r *UserRepository) FindAllUsers(ctx context.Context) ([]domain.User, error) {
  // Implementation goes here
  return []domain.User{}, nil
}
