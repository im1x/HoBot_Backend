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

func generateToken(user model.UserDto, secret string, expHour time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"id":           user.Id,
		"is_confirmed": user.IsConfirmed,
		"exp":          time.Now().Add(time.Hour * expHour).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))

	return t, err
}
func GenerateTokens(user model.UserDto) (string, string) {
	accessToken, err := generateToken(user, os.Getenv("JWT_ACCESS_SECRET"), 600)
	refreshToken, err := generateToken(user, os.Getenv("JWT_REFRESH_SECRET"), 1440)

	if err != nil {
		log.Error("Generate token error")
	}

	return accessToken, refreshToken
}

func isTokenValid(tokenString, secret string) (*model.UserDto, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.UserDto{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.UserDto); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func ValidateAccessToken(token string) (*model.UserDto, error) {
	return isTokenValid(token, os.Getenv("JWT_ACCESS_SECRET"))
}

func validateRefreshToken(token string) (*model.UserDto, error) {
	return isTokenValid(token, os.Getenv("JWT_REFRESH_SECRET"))
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

func removeToken(refreshToken string) error {
	one, err := DB.GetCollection(DB.Tokens).DeleteOne(ctx, bson.M{"refresh_token": refreshToken})
	if err != nil || one.DeletedCount == 0 {
		return err
	}
	return nil
}
func findToken(token string) (*model.Token, error) {
	var tokenDB model.Token
	err := DB.GetCollection(DB.Tokens).FindOne(ctx, bson.M{"refresh_token": token}).Decode(&tokenDB)
	if err != nil {
		return nil, err
	}
	return &tokenDB, nil
}
