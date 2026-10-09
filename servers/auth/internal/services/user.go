package services

import (
	"auth/internal/utils"
	"context"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (us *UserService) CreateUser(ctx context.Context, name, email string) (string, error) {
	userID, err := utils.GenerateUUID()

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (us *UserService) DeleteUser(ctx context.Context, userID string) error {
	return nil
}
