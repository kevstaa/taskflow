package service

import (
	"context"

	"github.com/kevstaa/taskflow/internal/model"
)

type UserService struct {
	userRepo UserRepo
}

func NewUserService(UserRepo UserRepo) *UserService {
	return &UserService{userRepo: UserRepo}
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id, username, email string) (*model.User, error) {
	return s.userRepo.Update(ctx, id, username, email)
}
