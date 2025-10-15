package user

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

type userRepository struct {
	col *mongo.Collection
}

func NewUserRepository(client *mongodb.Client) Repository {
	return &userRepository{
		col: client.GetCollection(mongodb.Users),
	}
}

func (r *userRepository) InsertOrUpdateUser(ctx context.Context, user model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	log.Info("VKPL: Inserting or updating user")
	_, err := r.col.UpdateByID(ctx, user.Id, bson.M{"$set": user}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		log.Error("Error while inserting or updating user:", err)
		return err
	}
	return nil
}

func (r *userRepository) GetUser(ctx context.Context, id string) (model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user model.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return user, err
}

func (r *userRepository) IsUserAlreadyExist(ctx context.Context, id string) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	candidate := r.col.FindOne(ctx, bson.M{"_id": id})
	return errors.Is(candidate.Err(), mongo.ErrNoDocuments)
}

func (r *userRepository) GetAndDeleteUser(ctx context.Context, id string) (model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user model.User
	err := r.col.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&user)

	return user, err
}

func (r *userRepository) GetUserIdByWs(ctx context.Context, ws string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user model.User
	err := r.col.FindOne(ctx, bson.M{"channel_ws": ws}).Decode(&user)
	return user.Id, err
}
