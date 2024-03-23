package user

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	settingsService "HoBot_Backend/pkg/service/settings"
	tokenService "HoBot_Backend/pkg/service/token"
	"HoBot_Backend/pkg/service/vkplay"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
)

var ctxParent = context.Background()

/*func Registration(ctx context.Context, user model.User) (*model.UserData, error) {
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
}*/

func GetCurrentUser(ctx context.Context, id string) (model.User, error) {
	var user model.User
	err := DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return user, err
}

func Logout(ctx context.Context, refreshToken string) error {
	return tokenService.RemoveToken(ctx, refreshToken)
}

func RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	userFromToken, err := tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	token, _ := tokenService.FindToken(ctx, refreshToken)
	if err != nil || token == nil {
		return "", "", fiber.NewError(fiber.StatusUnauthorized)
	}

	var userDb model.User
	err = DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"_id": userFromToken.Id}).Decode(&userDb)
	if err != nil {
		return "", "", err
	}

	accessToken, refreshToken := tokenService.GenerateTokens(userDb.Id)
	err = tokenService.SaveToken(ctx, userDb.Id, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, err
}

func isUserAlreadyExist(ctx context.Context, id string) bool {
	candidate := DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"_id": id})
	if errors.Is(candidate.Err(), mongo.ErrNoDocuments) {
		return true
	}
	return false
}

func LoginVkpl(ctx context.Context, currentUser model.CurrentUserVkpl) (string, error) {
	isNewUser := isUserAlreadyExist(ctx, strconv.Itoa(currentUser.Data.User.ID))
	user := model.User{
		Id:        strconv.Itoa(currentUser.Data.User.ID),
		Nick:      currentUser.Data.User.Nick,
		Channel:   currentUser.Data.Channel.Url,
		AvatarURL: currentUser.Data.User.AvatarURL + "&croped=1&mh=80&mw=80",
	}
	err := vkplay.InsertOrUpdateUser(ctx, user)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if isNewUser {
		err := vkplay.AddUserToWs(user.Id)
		if err != nil {
			log.Error(err)
			return "", err
		}
		err = settingsService.AddDefaultSettingsForUser(ctx, user.Id)
		if err != nil {
			log.Error(err)
			return "", err
		}
	}

	refreshToken := tokenService.GenerateRefreshToken(user.Id)
	err = tokenService.SaveToken(ctx, user.Id, refreshToken)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func WipeUser(ctx context.Context, id string) error {
	_, err := DB.GetCollection(DB.Users).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Error("Error while wiping user [Users]:", err)
		return err
	}

	_, err = DB.GetCollection(DB.UserSettings).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Error("Error while wiping user [UserSettings]:", err)
		return err
	}

	_, err = DB.GetCollection(DB.SongRequests).DeleteMany(ctx, bson.M{"channel_id": id})
	if err != nil {
		log.Error("Error while wiping user [SongRequests]:", err)
		return err
	}

	err = vkplay.RemoveUserFromWs(id)
	if err != nil {
		log.Error("Error while wiping user [WS]:", err)
		return err
	}

	return nil
}
