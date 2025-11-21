package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"database/sql"
	"errors"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAll() ([]model.User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetByUsername(username string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
       		if err == sql.ErrNoRows {
		        return nil, errors.New("users not found")
	        }
        	return nil, err
	}
	return user, nil
}

func (s *userService) GetByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("users not found")
		}
		return nil, err
	}
	return user, nil
}
