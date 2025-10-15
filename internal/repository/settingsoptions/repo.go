package settingsoptions

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	GetCommandDescription(ctx context.Context) (model.CommandsDescription, error)
	GetCommandList(ctx context.Context) (model.CommandList, error)
}
