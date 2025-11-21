package service

import "cruder/internal/repository"

type Service struct {
	Users UserService
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Users: NewUserService(repos.Users),
	}
}
