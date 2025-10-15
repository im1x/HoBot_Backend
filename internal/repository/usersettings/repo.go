package usersettings

import (
	"HoBot_Backend/internal/model"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	GetCommands(ctx context.Context) model.ChannelCommands
	GetAll(ctx context.Context) (*mongo.Cursor, error)
	UpdateSongRequestsSettingsByID(ctx context.Context, userId string, songReqSettings model.SongRequestsSettings) (*mongo.UpdateResult, error)
	GetDefaultSettings(ctx context.Context) (model.UserSettings, error)
	InsertOne(ctx context.Context, userSettings model.UserSettings) (*mongo.InsertOneResult, error)
	UpdateAliasesByID(ctx context.Context, userId string, aliases map[string]model.CmdDetails) (*mongo.UpdateResult, error)
	SaveVolume(ctx context.Context, userId string, volume int) error
	GetVolume(ctx context.Context, userId string) (int, error)
	ChangeVolumeBy(userId string, volume int) (int, error)
	DeleteUserSettings(ctx context.Context, userId string) error
}
