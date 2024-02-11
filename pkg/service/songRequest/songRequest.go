package songRequest

import (
	DB "HoBot_Backend/pkg/mongo"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
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

func GetPlaylist(ctxReq context.Context, streamer string) ([]SongRequest, error) {
	ctx, cancel := context.WithTimeout(ctxReq, 3*time.Second)
	var playlist []SongRequest
	defer cancel()
	cursor, err := DB.GetCollection(DB.SongRequests).Find(ctx, bson.M{"channel_id": streamer})
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

func GetUserIdByName(streamer string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.vkplay.live/v1/blog/%s/public_video_stream/chat/user/", streamer), nil)
	if err != nil {
		log.Error("Error while getting streamer id by name:", err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while getting streamer id by name:", err)
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
		log.Error("Error while reading streamer info:", err)
		return "", nil
	}

	err = json.Unmarshal(b, &streamerInfo)
	if err != nil {
		return "Error while unmarshal streamer info:", err
	}

	return strconv.Itoa(streamerInfo.Data.Owner.ID), nil

}
