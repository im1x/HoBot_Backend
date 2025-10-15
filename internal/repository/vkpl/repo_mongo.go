package vkpl

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type vkplRepository struct {
	col *mongo.Collection
}

func NewVkplRepository(client *mongodb.Client) Repository {
	return &vkplRepository{
		col: client.GetCollection(mongodb.Vkpl),
	}
}

func (r *vkplRepository) SaveAuth(ctx context.Context, auth model.AuthResponse) error {
	log.Info("VKPL: Saving vkplay auth to DB")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": "auth"}, auth)
	if err != nil {
		log.Error("Error while inserting vkplay auth:", err)
		return err
	}
	return nil
}

func (r *vkplRepository) GetAuth(ctx context.Context) (model.AuthResponse, error) {
	log.Info("VKPL: Getting vkplay auth from DB")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var auth model.AuthResponse
	err := r.col.FindOne(ctx, bson.M{"_id": "auth"}).Decode(&auth)
	if err != nil {
		log.Error("Error while getting vkplay auth:", err)
		return model.AuthResponse{}, err
	}
	return auth, nil
}
