package repository

import (
	"context"
	"cruder/internal/model"
	"database/sql"

	"log"
)

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	GetByUUID(uuid string) (*model.User, error) // Task3
	Create(user *model.User) error              // Task3
	Update(uuid string, user *model.User) error // Task3
	Delete(uuid string) error                   // Task3
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.QueryContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id int64) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByUUID(uuid string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(),
		`SELECT id, uuid, username, email, full_name FROM users WHERE uuid = $1`, uuid).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.QueryRowContext(context.Background(),
		`INSERT INTO users (username, email, full_name) VALUES ($1, $2, $3) RETURNING id, uuid`,
		user.Username, user.Email, user.FullName).
		Scan(&user.ID, &user.UUID)
}

func (r *userRepository) Update(uuid string, user *model.User) error {
	_, err := r.db.ExecContext(context.Background(),
		`UPDATE users SET username = $1, email = $2, full_name = $3 WHERE uuid = $4`,
		user.Username, user.Email, user.FullName, uuid)
	return err
}

func (r *userRepository) Delete(uuid string) error {
	result, err := r.db.ExecContext(context.Background(),
		`DELETE FROM users WHERE uuid = $1`, uuid)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
