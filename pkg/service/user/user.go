package user

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	tokenService "HoBot_Backend/pkg/service/token"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var ctxParent = context.Background()

func Registration(ctx context.Context, user model.User) (*model.UserData, error) {
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

	// --- TEMP ---
	user.Id = "18591758"
	res, err := colUser.InsertOne(ctx, user)
	userDto := user.ToUserDto()
	// --- TEMP ---
	userDto.Id = res.InsertedID.(string)
	//userDto.Id = res.InsertedID.(primitive.ObjectID)

	accessToken, refreshToken := tokenService.GenerateTokens(userDto)
	err = tokenService.SaveToken(ctx, res.InsertedID, refreshToken)
	if err != nil {
		return nil, err
	}

	resData := model.UserData{AccessToken: accessToken, RefreshToken: refreshToken, User: userDto}

	return &resData, err
}

func Login(ctx context.Context, user model.User) (*model.UserData, error) {
	if user.Login == "" || user.Password == "" {
		return nil, fiber.NewError(fiber.StatusConflict, "login or password is empty")
	}

	var userDb = model.User{}
	err := DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"login": user.Login}).Decode(&userDb)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "login or password incorrect")
	}
	err = bcrypt.CompareHashAndPassword([]byte(userDb.Password), []byte(user.Password+user.Login))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "login or password incorrect2")
	}

	userDto := userDb.ToUserDto()

	accessToken, refreshToken := tokenService.GenerateTokens(userDto)
	err = tokenService.SaveToken(ctx, userDb.Id, refreshToken)
	if err != nil {
		return nil, err
	}

	resData := model.UserData{AccessToken: accessToken, RefreshToken: refreshToken, User: userDto}

	return &resData, err
}

func Logout(ctx context.Context, refreshToken string) error {
	return tokenService.RemoveToken(ctx, refreshToken)
}

func RefreshToken(ctx context.Context, refreshToken string) (*model.UserData, error) {
	userFromToken, err := tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	token, _ := tokenService.FindToken(ctx, refreshToken)
	if err != nil || token == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized)
	}

	var userDb model.User
	err = DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"_id": userFromToken.Id}).Decode(&userDb)
	if err != nil {
		return nil, err
	}

	userDto := userDb.ToUserDto()

	accessToken, refreshToken := tokenService.GenerateTokens(userDto)
	err = tokenService.SaveToken(ctx, userDto.Id, refreshToken)
	if err != nil {
		return nil, err
	}

	resData := model.UserData{AccessToken: accessToken, RefreshToken: refreshToken, User: userDto}

	return &resData, err
}
