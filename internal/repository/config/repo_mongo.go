package config

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type configRepository struct {
	col *mongo.Collection
}

func NewConfigRepository(client *mongodb.Client) Repository {
	return &configRepository{
		col: client.GetCollection(mongodb.Config),
	}
}

func (r *configRepository) GetWsChannels(ctx context.Context) model.Config {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var config model.Config
	err := r.col.FindOne(ctx, bson.M{"_id": "ws"}).Decode(&config)
	if err != nil {
		log.Error("Error while getting channels:", err)
		return model.Config{}
	}
	return config
}

func (r *configRepository) SaveWsChannels(ctx context.Context, config model.Config) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.col.UpdateByID(ctx, "ws", bson.M{"$set": config})
	if err != nil {
		log.Error("Error while saving channels:", err)
		return err
	}
	return nil
}
