package statistics

import (
	DB "HoBot_Backend/pkg/mongo"
	"context"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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

func IncField(userId string, fieldName UpdateName) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		filter := bson.M{"_id": userId}
		update := bson.M{"$inc": bson.M{string(fieldName): 1}}
		opt := options.Update().SetUpsert(true)

		_, err := DB.GetCollection(DB.Statistics).UpdateOne(ctx, filter, update, opt)
		if err != nil {
			log.Errorf("Error while updating statistics. Field: %s. Error: %s", fieldName, err)
		}
	}()
}
