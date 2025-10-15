package config

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	GetWsChannels(ctx context.Context) model.Config
	SaveWsChannels(ctx context.Context, config model.Config) error
}
