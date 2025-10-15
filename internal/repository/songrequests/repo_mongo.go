package songrequests

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type songRequestsRepository struct {
	col *mongo.Collection
}

func NewSongRequestsRepository(client *mongodb.Client) Repository {
	return &songRequestsRepository{
		col: client.GetCollection(mongodb.SongRequests),
	}
}

func (r *songRequestsRepository) AddSongRequestToDB(ctx context.Context, songRequest model.SongRequest) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := r.col.InsertOne(ctx, songRequest)
	if err != nil {
		log.Error("Error while inserting song request:", err)
		return bson.NilObjectID, err
	}

	insertedID := res.InsertedID.(bson.ObjectID)

	return insertedID, nil
}

func (r *songRequestsRepository) GetPlaylist(ctx context.Context, user string) ([]model.SongRequest, error) {
	var playlist []model.SongRequest
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cursor, err := r.col.Find(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while finding song request:", err)
		return nil, err
	}

	for cursor.Next(ctx) {
		var song model.SongRequest
		if err := cursor.Decode(&song); err != nil {
			log.Error("Error decoding song request:", err)
			continue
		}
		playlist = append(playlist, song)
	}

	if err := cursor.Err(); err != nil {
		log.Error("Error iterating over cursor:", err)
		return nil, err
	}

	return playlist, nil
}

func (r *songRequestsRepository) IsPlaylistFull(ctx context.Context, user string) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	playlistCount, err := r.col.CountDocuments(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while getting playlist count:", err)
		return false
	}

	if playlistCount >= 30 {
		return true
	}

	return false
}

func (r *songRequestsRepository) RemoveAllSongs(ctx context.Context, channelId string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.col.DeleteMany(ctx, bson.M{"channel_id": channelId})
	if err != nil {
		log.Error("Error while deleting playlist:", err)
		return err
	}
	return nil
}

func (r *songRequestsRepository) RemoveSong(ctx context.Context, channelId, songId string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	objectId, err := bson.ObjectIDFromHex(songId)
	if err != nil {
		return err
	}

	res, err := r.col.DeleteOne(ctx, bson.M{"channel_id": channelId, "_id": objectId})
	if err != nil || res.DeletedCount != 1 {
		log.Error("Error while deleting song:", err)
		return errors.New("error while deleting song")
	}
	return nil
}

func (r *songRequestsRepository) DeleteSongByYtId(ctx context.Context, channelId, ytId string) (model.SongRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var song model.SongRequest
	err := r.col.FindOneAndDelete(ctx, bson.M{"channel_id": channelId, "yt_id": ytId}).Decode(&song)
	if err != nil {
		log.Error("Error while deleting song:", err)
		return model.SongRequest{}, errors.New("error while deleting song")
	}

	return song, nil
}

func (r *songRequestsRepository) GetCurrentSong(ctx context.Context, channelId string) (model.SongRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var curSong model.SongRequest
	filter := bson.M{"channel_id": channelId}
	opt := options.FindOne().SetSort(bson.D{{"_id", 1}})
	res := r.col.FindOne(ctx, filter, opt)

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return model.SongRequest{}, nil
		}
		log.Error("Error while getting current song:", res.Err())
		return model.SongRequest{}, res.Err()
	}

	err := res.Decode(&curSong)
	if err != nil {
		log.Error("Error decoding current song:", err)
		return model.SongRequest{}, err
	}

	return curSong, nil
}

func (r *songRequestsRepository) CountSongsByUser(ctx context.Context, channelId string, userName string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := r.col.CountDocuments(ctx, bson.M{"channel_id": channelId, "by": userName})
	if err != nil {
		log.Error("Error while counting songs by user:", err)
		return 0, err
	}
	return int(res), nil
}

func (r *songRequestsRepository) SkipSong(ctx context.Context, channelId string) (model.SongRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var song model.SongRequest

	filter := bson.M{"channel_id": channelId}
	opt := options.FindOneAndDelete().SetSort(bson.D{{"_id", 1}})
	err := r.col.FindOneAndDelete(ctx, filter, opt).Decode(&song)
	return song, err
}

func (r *songRequestsRepository) DeleteAllSongRequests(ctx context.Context, channelId string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.col.DeleteMany(ctx, bson.M{"channel_id": channelId})
	return err
}
