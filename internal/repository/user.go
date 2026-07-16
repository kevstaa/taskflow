package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kevstaa/taskflow/db"
	"github.com/kevstaa/taskflow/internal/model"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, username, email, passwordHash string) (*model.User, error) {
	dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}
	return toModelUser(dbUser), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toModelUser(dbUser), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var uuid pgtype.UUID
	uuid.Scan(id)

	dbUser, err := r.queries.GetUserByID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return toModelUser(dbUser), nil
}

func (r *UserRepository) Update(ctx context.Context, id, username, email string) (*model.User, error) {
	var uuid pgtype.UUID
	uuid.Scan(id)

	dbUser, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       uuid,
		Username: username,
		Email:    email,
	})
	if err != nil {
		return nil, err
	}
	return toModelUser(dbUser), nil
}

// helper convert db.User -> model.User
func toModelUser(u db.User) *model.User {
	return &model.User{
		ID:           u.ID.String(),
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}
