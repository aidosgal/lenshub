package service

import (

	"github.com/aidosgal/lenshub/internal/model"
)

type UserRepository interface {
    CreateUser(user model.User) error
    GetUserByChatID(string) (*model.User, error)
    GetUsersBySpecialization(specialization string) ([]model.User, error)
}

type UserService struct {
    repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{
        repository: repository,
	}
}

func (s *UserService) CreateUser(user model.User) error {
	return s.repository.CreateUser(user)
}

func (s *UserService) GetUserByChatID(chatID string) (*model.User, error) {
    return s.repository.GetUserByChatID(chatID)
}

func (s *UserService) GetUsersBySpecialization(specialization string) ([]model.User, error) {
    return s.repository.GetUsersBySpecialization(specialization)
}
