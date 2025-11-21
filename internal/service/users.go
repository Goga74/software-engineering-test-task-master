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
	GetByUUID(uuid string) (*model.User, error) // Task3
	Create(user *model.User) error              // Task3
	Update(uuid string, user *model.User) error // Task3
	Delete(uuid string) error                   // Task3
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
			return nil, errors.New("users not found") // Task2
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("users not found") // Task2
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByUUID(uuid string) (*model.User, error) {
	user, err := s.repo.GetByUUID(uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("users not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) Create(user *model.User) error {
	// validate uniq username
	existingUser, _ := s.repo.GetByUsername(user.Username)
	if existingUser != nil {
		return errors.New("username already exists")
	}

	return s.repo.Create(user)
}

func (s *userService) Update(uuid string, user *model.User) error {
	// check that user exists
	existingUser, err := s.repo.GetByUUID(uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("users not found")
		}
		return err
	}

	// check that user exists
	if user.Username != existingUser.Username {
		userByName, _ := s.repo.GetByUsername(user.Username)
		if userByName != nil && userByName.UUID != uuid {
			return errors.New("username already exists")
		}
	}

	return s.repo.Update(uuid, user)
}

func (s *userService) Delete(uuid string) error {
	err := s.repo.Delete(uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("users not found")
		}
		return err
	}
	return nil
}
