package usersettings

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

type userSettingsRepository struct {
	col *mongo.Collection
}

func NewUserSettingsRepository(client *mongodb.Client) Repository {
	return &userSettingsRepository{
		col: client.GetCollection(mongodb.UserSettings),
	}
}

func (r *userSettingsRepository) GetCommands(ctx context.Context) model.ChannelCommands {
	var cmds model.ChannelCommands
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Set up the aggregation pipeline
	pipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", nil},
			{"channels", bson.D{{"$push", bson.D{
				{"k", "$_id"},
				{"v", bson.D{{"aliases", "$aliases"}}},
			}}}},
		}}},
		{{"$replaceRoot", bson.D{
			{"newRoot", bson.D{
				{"_id", "commands"},
				{"channels", bson.D{{"$arrayToObject", "$channels"}}},
			}},
		}}},
	}

	// Execute the aggregation
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Error while aggregating:", err)
	}
	defer cursor.Close(ctx)

	// Iterate over the result
	if cursor.Next(ctx) {
		err := cursor.Decode(&cmds)
		if err != nil {
			log.Error("Error while decoding:", err)
		}
	}

	return cmds
}

func (r *userSettingsRepository) GetAll(ctx context.Context) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.col.Find(ctx, bson.M{})
}

func (r *userSettingsRepository) UpdateSongRequestsSettingsByID(ctx context.Context, userId string, songReqSettings model.SongRequestsSettings) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.col.UpdateByID(ctx, userId, bson.M{"$set": bson.M{"song_requests": songReqSettings}})
}

func (r *userSettingsRepository) GetDefaultSettings(ctx context.Context) (model.UserSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var settings model.UserSettings
	err := r.col.FindOne(ctx, bson.M{"_id": "default"}).Decode(&settings)
	return settings, err
}

func (r *userSettingsRepository) InsertOne(ctx context.Context, userSettings model.UserSettings) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.col.InsertOne(ctx, userSettings)
}

func (r *userSettingsRepository) UpdateAliasesByID(ctx context.Context, userId string, aliases map[string]model.CmdDetails) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.col.UpdateByID(ctx, userId, bson.M{"$set": bson.M{"aliases": aliases}})
}

func (r *userSettingsRepository) SaveVolume(ctx context.Context, userId string, volume int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.col.UpdateByID(ctx, userId, bson.M{"$set": bson.M{"volume": volume}})
	if err != nil {
		log.Error("Error while updating volume:", err)
		return err
	}
	return nil
}

func (r *userSettingsRepository) GetVolume(ctx context.Context, userId string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var userSettings model.UserSettings
	filter := bson.D{{"_id", userId}}
	opts := options.FindOne().SetProjection(bson.D{{"volume", 1}})
	err := r.col.FindOne(ctx, filter, opts).Decode(&userSettings)
	if err != nil {
		log.Error("Error while getting volume:", err)
		return 0, err
	}
	return userSettings.Volume, nil
}

func (r *userSettingsRepository) ChangeVolumeBy(userId string, volume int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vol, err := r.GetVolume(ctx, userId)
	if err != nil {
		return 0, err
	}

	vol += volume
	vol = max(0, min(vol, 100))

	err = r.SaveVolume(ctx, userId, vol)
	if err != nil {
		return 0, err
	}

	return vol, nil
}

func (r *userSettingsRepository) DeleteUserSettings(ctx context.Context, userId string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.DeleteOne(ctx, bson.M{"_id": userId})

	return err
}
