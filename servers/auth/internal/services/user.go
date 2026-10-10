package services

import (
	usersv1 "auth/internal/gen/users/v1"
	"context"
)

type UserService struct {
	client usersv1.UserServiceClient
}

func NewUserService(client usersv1.UserServiceClient) *UserService {
	return &UserService{client: client}
}

func (us *UserService) CreateUser(ctx context.Context, name, email string) (string, error) {
	r, err := us.client.CreateUser(ctx, &usersv1.CreateUserRequest{
		Email: email,
		Name:  name,
	})

	if err != nil {
		return "", err
	}

	return r.GetUserId(), nil
}

func (us *UserService) DeleteUser(ctx context.Context, userID string) error {
	_, err := us.client.DeleteUser(ctx, &usersv1.DeleteUserRequest{UserId: userID})
	return err
}
