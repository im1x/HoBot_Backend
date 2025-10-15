package songrequests

import (
	"HoBot_Backend/internal/model"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository interface {
	AddSongRequestToDB(ctx context.Context, songRequest model.SongRequest) (bson.ObjectID, error)
	GetPlaylist(ctx context.Context, user string) ([]model.SongRequest, error)
	IsPlaylistFull(ctx context.Context, user string) bool
	RemoveAllSongs(ctx context.Context, channelId string) error
	RemoveSong(ctx context.Context, channelId, songId string) error
	DeleteSongByYtId(ctx context.Context, channelId, ytId string) (model.SongRequest, error)
	GetCurrentSong(ctx context.Context, channelId string) (model.SongRequest, error)
	CountSongsByUser(ctx context.Context, channelId string, userName string) (int, error)
	SkipSong(ctx context.Context, channelId string) (model.SongRequest, error)
	DeleteAllSongRequests(ctx context.Context, channelId string) error
}
