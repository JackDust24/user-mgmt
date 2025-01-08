package data

import (
	"context"
	"database/sql"
	"user-mgmt/pkg/models"
	"user-mgmt/pkg/repository"
)

type UserFetcher interface {
	GetUser(ctx context.Context) (models.User, error)
}

type DBUserFetcher struct {
	DB     *sql.DB
	UserID string
}

func (f *DBUserFetcher) GetUser(ctx context.Context) (models.User, error) {
	// Implement your database fetch logic here
	return repository.GetUserById(f.DB, f.UserID)
}
