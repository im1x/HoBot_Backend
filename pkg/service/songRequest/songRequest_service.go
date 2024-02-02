package songRequest

import (
	DB "HoBot_Backend/pkg/mongo"
	"context"
	"github.com/gofiber/fiber/v2/log"
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
