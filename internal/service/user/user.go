package user

import (
	"HoBot_Backend/internal/model"
	repoSongRequests "HoBot_Backend/internal/repository/songrequests"
	repoSongRequestsHistory "HoBot_Backend/internal/repository/songrequestshistory"
	repoToken "HoBot_Backend/internal/repository/token"
	repoUser "HoBot_Backend/internal/repository/user"
	repoUserSettings "HoBot_Backend/internal/repository/usersettings"
	"HoBot_Backend/internal/service/chat"
	"HoBot_Backend/internal/service/settings"
	tokenService "HoBot_Backend/internal/service/token"
	"HoBot_Backend/internal/service/vkplay"
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type UserService struct {
	ctxApp                  context.Context
	userRepo                repoUser.Repository
	userSettingsRepo        repoUserSettings.Repository
	songRequestsRepo        repoSongRequests.Repository
	songRequestsHistoryRepo repoSongRequestsHistory.Repository
	tokenRepo               repoToken.Repository
	vkplService             *vkplay.VkplService
	settingsService         *settings.SettingsService
	chatService             *chat.ChatService
}

func NewUserService(
	ctx context.Context,
	userRepo repoUser.Repository,
	userSettingsRepo repoUserSettings.Repository,
	songRequestsRepo repoSongRequests.Repository,
	songRequestsHistoryRepo repoSongRequestsHistory.Repository,
	tokenRepo repoToken.Repository,
	vkplService *vkplay.VkplService,
	settingsService *settings.SettingsService,
	chatService *chat.ChatService,
) *UserService {
	return &UserService{
		ctxApp:                  ctx,
		userRepo:                userRepo,
		userSettingsRepo:        userSettingsRepo,
		songRequestsRepo:        songRequestsRepo,
		songRequestsHistoryRepo: songRequestsHistoryRepo,
		tokenRepo:               tokenRepo,
		vkplService:             vkplService,
		settingsService:         settingsService,
		chatService:             chatService,
	}
}

func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.RemoveToken(ctx, refreshToken)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	userFromToken, err := tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	token, err := s.tokenRepo.FindToken(ctx, refreshToken)
	if err != nil || token == nil {
		return "", "", fiber.NewError(fiber.StatusUnauthorized)
	}

	userDb, err := s.userRepo.GetUser(ctx, userFromToken.Id)
	if err != nil {
		return "", "", err
	}

	accessToken, refreshToken := tokenService.GenerateTokens(userDb.Id)
	err = s.tokenRepo.SaveToken(ctx, userDb.Id, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, err
}

func (s *UserService) LoginVkpl(ctx context.Context, currentUser model.CurrentUserVkpl) (string, error) {
	isNewUser := s.userRepo.IsUserAlreadyExist(ctx, strconv.Itoa(currentUser.Data.User.ID))
	channelInfo, err := vkplay.GetChannelInfo(currentUser.Data.Channel.Url)
	if err != nil {
		return "", err
	}

	channelWs := strings.Split(channelInfo.Data.Channel.WebSocketChannels.Chat, ":")[1]

	user := model.User{
		Id:        strconv.Itoa(currentUser.Data.User.ID),
		Nick:      currentUser.Data.User.Nick,
		Channel:   currentUser.Data.Channel.Url,
		ChannelWS: channelWs,
		AvatarURL: currentUser.Data.User.AvatarURL + "&croped=1&mh=80&mw=80",
	}

	err = s.userRepo.InsertOrUpdateUser(ctx, user)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if isNewUser {
		err := vkplay.FollowUnfollowChannel(user.Channel, true, s.vkplService.GetVkplToken())
		if err != nil {
			log.Error(err)
			return "", err
		}

		err = s.chatService.AddUserToWs(user)
		if err != nil {
			log.Error(err)
			return "", err
		}

		err = s.settingsService.AddDefaultSettingsForUser(ctx, user)
		if err != nil {
			log.Error(err)
			return "", err
		}
	}

	refreshToken := tokenService.GenerateRefreshToken(user.Id)
	err = s.tokenRepo.SaveToken(ctx, user.Id, refreshToken)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *UserService) WipeUser(ctx context.Context, id string) error {
	user, err := s.userRepo.GetAndDeleteUser(ctx, id)
	if err != nil {
		log.Error("Error while wiping user [Users]:", err)
		return err
	}

	err = s.userSettingsRepo.DeleteUserSettings(ctx, id)
	if err != nil {
		log.Error("Error while wiping user [UserSettings]:", err)
		return err
	}

	err = s.songRequestsRepo.DeleteAllSongRequests(ctx, id)
	if err != nil {
		log.Error("Error while wiping user [SongRequests]:", err)
		return err
	}

	err = s.songRequestsHistoryRepo.DeleteAllSongRequestsHistory(ctx, id)
	if err != nil {
		log.Error("Error while wiping user [SongRequestsHistory]:", err)
		return err
	}

	err = s.chatService.RemoveUserFromWs(user)
	if err != nil {
		log.Error("Error while wiping user [WS]:", err)
		return err
	}

	err = vkplay.FollowUnfollowChannel(user.Channel, false, s.vkplService.GetVkplToken())
	if err != nil {
		log.Error("Error while unfollowing channel:", err)
		return err
	}

	delete(s.vkplService.ChannelsCommands.Channels, id)
	delete(s.settingsService.UsersSettings, id)

	return nil
}
