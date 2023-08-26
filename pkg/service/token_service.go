package service

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"errors"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
)

func generateToken(user model.User, secret string, expHour time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":      user.Id,
		"is_confirmed": user.IsConfirmed,
		"exp":          time.Now().Add(time.Hour * expHour).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))

	return t, err
}
func GenerateTokens(user model.User) (string, string) {
	accessToken, err := generateToken(user, os.Getenv("JWT_ACCESS_SECRET"), 6)
	refreshToken, err := generateToken(user, os.Getenv("JWT_REFRESH_SECRET"), 24)

	if err != nil {
		log.Error("Generate token error")
	}

	return accessToken, refreshToken
}

func saveToken(uid interface{}, refreshToken string) error {
	colToken := DB.GetCollection(DB.Tokens)
	var err error
	filter := bson.M{"user_id": uid}

	res := colToken.FindOne(ctx, filter)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			_, err = colToken.InsertOne(ctx, bson.M{"user_id": uid, "refresh_token": refreshToken})
			return nil
		}
		log.Error("Error while querying existing token:", err)
		return err
	}

	_, err = colToken.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"refresh_token": refreshToken}})
	if err != nil {
		log.Error("Error while updating token:", err)
	}

	return err
}
