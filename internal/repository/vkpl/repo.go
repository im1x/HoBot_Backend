package vkpl

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	SaveAuth(ctx context.Context, auth model.AuthResponse) error
	GetAuth(ctx context.Context) (model.AuthResponse, error)
}
