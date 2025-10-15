package statistics

import (
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UpdateName string

const (
	SongRequestsAdd                UpdateName = "song_requests_add"
	SongRequestsSkipByCommand                 = "song_requests_skip_by_command"
	SongRequestsSkipByUsers                   = "song_requests_skip_by_users"
	SongRequestsShowPublicPlaylist            = "song_requests_show_public_playlist"
	SongRequestsPlayed                        = "song_requests_played"
	PrintTextByCommand                        = "print_text_by_command"
	Voting                                    = "voting"
	Rating                                    = "rating"
)

type statisticsRepository struct {
	col *mongo.Collection
}

func NewStatisticsRepository(client *mongodb.Client) Repository {
	return &statisticsRepository{
		col: client.GetCollection(mongodb.Statistics),
	}
}

func (r *statisticsRepository) IncField(ctx context.Context, userId string, fieldName UpdateName) {
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		filter := bson.M{"_id": userId}
		update := bson.M{"$inc": bson.M{string(fieldName): 1}}
		opt := options.UpdateOne().SetUpsert(true)

		_, err := r.col.UpdateOne(ctx, filter, update, opt)
		if err != nil {
			log.Errorf("Error while updating statistics. Field: %s. Error: %s", fieldName, err)
		}
	}()
}
