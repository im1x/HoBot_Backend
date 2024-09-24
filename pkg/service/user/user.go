package user

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service/chat"
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

	token, err := tokenService.FindToken(ctx, refreshToken)
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
	return errors.Is(candidate.Err(), mongo.ErrNoDocuments)
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
		err := chat.AddUserToWs(user.Id)
		if err != nil {
			log.Error(err)
			return "", err
		}

		err = settingsService.AddDefaultSettingsForUser(ctx, user)
		if err != nil {
			log.Error(err)
			return "", err
		}

		err = vkplay.FollowUnfollowChannel(user.Channel, true)
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
	var user model.User
	err := DB.GetCollection(DB.Users).FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&user)
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

	err = chat.RemoveUserFromWs(id)
	if err != nil {
		log.Error("Error while wiping user [WS]:", err)
		return err
	}

	err = vkplay.FollowUnfollowChannel(user.Channel, false)
	if err != nil {
		log.Error("Error while unfollowing channel:", err)
		return err
	}

	delete(vkplay.ChannelsCommands.Channels, id)
	delete(settingsService.UsersSettings, id)

	return nil
}
