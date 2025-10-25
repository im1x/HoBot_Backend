package token

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type tokenRepository struct {
	col *mongo.Collection
}

func NewTokenRepository(client *mongodb.Client) Repository {
	return &tokenRepository{
		col: client.GetCollection(mongodb.Tokens),
	}
}
func (r *tokenRepository) SaveToken(ctx context.Context, uid string, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var err error
	filter := bson.M{"user_id": uid}

	res := r.col.FindOne(ctx, filter)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			_, err = r.col.InsertOne(ctx, bson.M{"user_id": uid, "refresh_token": refreshToken})
			if err != nil {
				log.Error("Error while inserting token:", err)
				return err
			}
			return nil
		}
		log.Error("Error while querying existing token:", err)
		return err
	}

	_, err = r.col.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"refresh_token": refreshToken}})
	if err != nil {
		log.Error("Error while updating token:", err)
	}

	return err
}

func (r *tokenRepository) RemoveToken(ctx context.Context, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	one, err := r.col.DeleteOne(ctx, bson.M{"refresh_token": refreshToken})
	if err != nil || one.DeletedCount == 0 {
		return err
	}
	return nil
}

func (r *tokenRepository) RemoveTokenByChannelId(ctx context.Context, channelId string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.col.DeleteOne(ctx, bson.M{"user_id": channelId})
	if err != nil {
		log.Error("Error while deleting token by channel id:", err)
		return err
	}
	return nil
}

func (r *tokenRepository) FindToken(ctx context.Context, token string) (*model.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var tokenDB model.Token
	err := r.col.FindOne(ctx, bson.M{"refresh_token": token}).Decode(&tokenDB)
	if err != nil {
		return nil, err
	}
	return &tokenDB, nil
}
