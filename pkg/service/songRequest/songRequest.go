package songRequest

import (
	DB "HoBot_Backend/pkg/mongo"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"strconv"
	"time"
)

func AddSongRequestToDB(songRequest SongRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.SongRequests).InsertOne(ctx, songRequest)
	if err != nil {
		log.Error("Error while inserting song request:", err)
		return err
	}
	return nil
}

func GetPlaylist(ctxReq context.Context, user string) ([]SongRequest, error) {
	var playlist []SongRequest
	ctx, cancel := context.WithTimeout(ctxReq, 3*time.Second)
	defer cancel()

	cursor, err := DB.GetCollection(DB.SongRequests).Find(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while finding song request:", err)
		return nil, err
	}

	for cursor.Next(ctx) {
		var song SongRequest
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

func GetUserIdByName(user string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.live.vkplay.ru/v1/blog/%s/public_video_stream/chat/user/", user), nil)
	if err != nil {
		log.Error("Error while getting user id by name:", err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while getting user id by name:", err)
	}
	defer resp.Body.Close()

	type StreamerInfo struct {
		Data struct {
			Owner struct {
				ID int `json:"id"`
			} `json:"owner"`
		} `json:"data"`
	}

	var streamerInfo StreamerInfo
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error while reading user info:", err)
		return "", nil
	}

	err = json.Unmarshal(b, &streamerInfo)
	if err != nil {
		return "Error while unmarshal user info:", err
	}

	return strconv.Itoa(streamerInfo.Data.Owner.ID), nil

}

func IsPlaylistFull(user string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	playlistCount, err := DB.GetCollection(DB.SongRequests).CountDocuments(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while getting playlist count:", err)
		return false
	}

	if playlistCount >= 30 {
		return true
	}

	return false
}

func SkipSong(channelId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"channel_id": channelId}
	opt := options.FindOneAndDelete().SetSort(bson.D{{"_id", 1}})
	result := DB.GetCollection(DB.SongRequests).FindOneAndDelete(ctx, filter, opt)
	if result.Err() != nil {
		log.Error("Error while deleting song request:", result.Err())
		return result.Err()
	}
	return nil
}

func RemoveAllSongs(channelId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.SongRequests).DeleteMany(ctx, bson.M{"channel_id": channelId})
	if err != nil {
		log.Error("Error while deleting playlist:", err)
		return err
	}
	return nil
}

func RemoveSong(channelId, songId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(songId)
	if err != nil {
		return err
	}

	res, err := DB.GetCollection(DB.SongRequests).DeleteOne(ctx, bson.M{"channel_id": channelId, "_id": objectId})
	if err != nil || res.DeletedCount != 1 {
		log.Error("Error while deleting song:", err)
		return errors.New("error while deleting song")
	}
	return nil
}

func GetCurrentSong(channelId string) (SongRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var curSong SongRequest
	filter := bson.M{"channel_id": channelId}
	opt := options.FindOne().SetSort(bson.D{{"_id", 1}})
	res := DB.GetCollection(DB.SongRequests).FindOne(ctx, filter, opt)

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return SongRequest{}, nil
		}
		log.Error("Error while getting current song:", res.Err())
		return SongRequest{}, res.Err()
	}

	err := res.Decode(&curSong)
	if err != nil {
		log.Error("Error decoding current song:", err)
		return SongRequest{}, err
	}

	return curSong, nil
}
