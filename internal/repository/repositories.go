package repository

import "database/sql"

type Repository struct {
	Users UserRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Users: NewUserRepository(db),
	}
}
