package settingsoptions

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type settingsOptionsRepository struct {
	col *mongo.Collection
}

func NewSettingsOptionsRepository(client *mongodb.Client) Repository {
	return &settingsOptionsRepository{
		col: client.GetCollection(mongodb.SettingsOptions),
	}
}

func (r *settingsOptionsRepository) GetCommandDescription(ctx context.Context) (model.CommandsDescription, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var commandsDescription model.CommandsDescription
	err := r.col.FindOne(ctx, bson.M{"_id": "commandsDescription"}).Decode(&commandsDescription)
	if err != nil {
		log.Error("Error while getting command description:", err)
		return model.CommandsDescription{}, err
	}
	return commandsDescription, nil
}

func (r *settingsOptionsRepository) GetCommandList(ctx context.Context) (model.CommandList, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var commandsList model.CommandList
	err := r.col.FindOne(ctx, bson.M{"_id": "commandsList"}).Decode(&commandsList)
	return commandsList, err

}
