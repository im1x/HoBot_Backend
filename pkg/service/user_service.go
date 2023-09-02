package service

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var ctx = context.TODO()

func Registration(user model.User) (*model.UserData, error) {
	if user.Login == "" || user.Password == "" {
		return nil, fiber.NewError(fiber.StatusConflict, "login or password is empty")
	}
	colUser := DB.GetCollection(DB.Users)

	candidate := colUser.FindOne(ctx, bson.M{"login": user.Login})
	if !errors.Is(candidate.Err(), mongo.ErrNoDocuments) {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "login already used")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password+user.Login), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	res, err := colUser.InsertOne(ctx, user)
	userDto := user.ToUserDto()
	userDto.Id = res.InsertedID.(primitive.ObjectID)

	accessToken, refreshToken := GenerateTokens(userDto)
	err = saveToken(res.InsertedID, refreshToken)
	if err != nil {
		return nil, err
	}

	resData := model.UserData{AccessToken: accessToken, RefreshToken: refreshToken, User: userDto}

	return &resData, err
}

func Login(user model.User) (*model.UserData, error) {
	if user.Login == "" || user.Password == "" {
		return nil, fiber.NewError(fiber.StatusConflict, "login or password is empty")
	}
	colUser := DB.GetCollection(DB.Users)

	var userDb = model.User{}
	err := colUser.FindOne(ctx, bson.M{"login": user.Login}).Decode(&userDb)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "login or password incorrect")
	}
	err = bcrypt.CompareHashAndPassword([]byte(userDb.Password), []byte(user.Password+user.Login))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "login or password incorrect2")
	}

	userDto := userDb.ToUserDto()

	accessToken, refreshToken := GenerateTokens(userDto)
	err = saveToken(userDb.Id, refreshToken)
	if err != nil {
		return nil, err
	}

	resData := model.UserData{AccessToken: accessToken, RefreshToken: refreshToken, User: userDto}

	return &resData, err
}

func Logout(refreshToken string) error {
	return removeToken(refreshToken)
}

func RefreshToken(refreshToken string) (*model.UserData, error) {
	colUser := DB.GetCollection(DB.Users)
	userFromToken, err := validateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	token, _ := findToken(refreshToken)
	if err != nil || token == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized)
	}

	var userDb model.User
	err = colUser.FindOne(ctx, bson.M{"_id": userFromToken.Id}).Decode(&userDb)
	if err != nil {
		return nil, err
	}

	userDto := userDb.ToUserDto()

	accessToken, refreshToken := GenerateTokens(userDto)
	err = saveToken(userDto.Id, refreshToken)
	if err != nil {
		return nil, err
	}

	resData := model.UserData{AccessToken: accessToken, RefreshToken: refreshToken, User: userDto}

	return &resData, err
}
