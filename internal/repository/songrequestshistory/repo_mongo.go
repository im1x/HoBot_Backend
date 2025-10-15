package songrequestshistory

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type songRequestsHistoryRepository struct {
	col *mongo.Collection
}

func NewSongRequestsHistoryRepository(client *mongodb.Client) Repository {
	return &songRequestsHistoryRepository{
		col: client.GetCollection(mongodb.SongRequestsHistory),
	}
}
func (r *songRequestsHistoryRepository) GetPlaylistHistory(ctx context.Context, user string) ([]model.SongRequest, error) {
	var playlistHistory []model.SongRequest
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cursor, err := r.col.Find(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while finding song requests history:", err)
		return nil, err
	}

	for cursor.Next(ctx) {
		var song model.SongRequest
		if err := cursor.Decode(&song); err != nil {
			log.Error("Error decoding song request history:", err)
			continue
		}
		playlistHistory = append(playlistHistory, song)
	}

	if err := cursor.Err(); err != nil {
		log.Error("Error iterating over cursor:", err)
		return nil, err
	}

	return playlistHistory, nil
}

func (r *songRequestsHistoryRepository) SaveSongRequestToHistory(song model.SongRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// count user documents in history
	count, err := r.col.CountDocuments(ctx, bson.M{"channel_id": song.ChannelId})
	if err != nil {
		log.Error("Error while counting user documents in history:", err)
		return err
	}

	if count >= 3 {
		// delete oldest document
		filter := bson.M{"channel_id": song.ChannelId}
		opt := options.FindOneAndDelete().SetSort(bson.D{{"_id", 1}})
		result := r.col.FindOneAndDelete(ctx, filter, opt)
		if result.Err() != nil {
			log.Error("Error while deleting oldest document from history:", result.Err())
			return result.Err()
		}
	}

	_, err = r.col.InsertOne(ctx, song)
	if err != nil {
		log.Error("Error while inserting song to history:", err)
		return err
	}

	return nil
}

func (r *songRequestsHistoryRepository) DeleteAllSongRequestsHistory(ctx context.Context, channelId string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.col.DeleteMany(ctx, bson.M{"channel_id": channelId})
	return err
}
