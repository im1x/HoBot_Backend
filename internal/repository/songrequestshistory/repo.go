package songrequestshistory

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	GetPlaylistHistory(ctx context.Context, user string) ([]model.SongRequest, error)
	SaveSongRequestToHistory(song model.SongRequest) error
	DeleteAllSongRequestsHistory(ctx context.Context, channelId string) error
}
