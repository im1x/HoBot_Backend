package user

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	InsertOrUpdateUser(ctx context.Context, user model.User) error
	GetUser(ctx context.Context, id string) (model.User, error)
	IsUserAlreadyExist(ctx context.Context, id string) bool
	GetAndDeleteUser(ctx context.Context, id string) (model.User, error)
	GetUserIdByWs(ctx context.Context, ws string) (string, error)
}
