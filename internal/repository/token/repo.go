package token

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	SaveToken(ctx context.Context, uid string, refreshToken string) error
	RemoveToken(ctx context.Context, refreshToken string) error
	RemoveTokenByChannelId(ctx context.Context, channelId string) error
	FindToken(ctx context.Context, token string) (*model.Token, error)
}
