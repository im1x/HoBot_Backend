package settings

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/service/vkplay"
	"context"
	"sync"
	"time"

	repoSettingsOptions "HoBot_Backend/internal/repository/settingsoptions"
	repoUserSettings "HoBot_Backend/internal/repository/usersettings"

	"github.com/gofiber/fiber/v2/log"
)

type SettingsService struct {
	ctxApp              context.Context
	UsersSettings       map[string]model.UserSettings
	userSettingsRepo    repoUserSettings.Repository
	settingsOptionsRepo repoSettingsOptions.Repository
	vkplService         vkplay.VkplService
}

func NewSettingsService(
	ctx context.Context,
	userSettingsRepo repoUserSettings.Repository,
	settingsOptionsRepo repoSettingsOptions.Repository,
	vkplService *vkplay.VkplService,
) *SettingsService {
	settingsService := &SettingsService{
		ctxApp:              ctx,
		UsersSettings:       make(map[string]model.UserSettings),
		userSettingsRepo:    userSettingsRepo,
		settingsOptionsRepo: settingsOptionsRepo,
		vkplService:         *vkplService,
	}
	settingsService.LoadSettings()
	return settingsService
}

func (s *SettingsService) LoadSettings() {
	ctx, cancel := context.WithTimeout(s.ctxApp, 5*time.Second)
	defer cancel()

	cursor, err := s.userSettingsRepo.GetAll(ctx)
	if err != nil {
		log.Error("Error while loading settings:", err)
		return
	}

	for cursor.Next(ctx) {
		var result model.UserSettings
		if err := cursor.Decode(&result); err != nil {
			log.Error("Error while decoding settings:", err)
			continue
		}
		result.Aliases = nil
		result.Volume = 0
		s.UsersSettings[result.Id] = result
	}
}

func (s *SettingsService) SaveSongRequestSettings(ctx context.Context, userId string, songReqSettings model.SongRequestsSettings) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	userSetting := s.UsersSettings[userId]

	if !userSetting.SongRequests.IsUsersSkipAllowed && songReqSettings.IsUsersSkipAllowed {
		s.AddUsersSkipCommands(ctx, userId)
		//songRequest.InitUsersSkipIfNeeded(userId)
	}

	userSetting.SongRequests = songReqSettings
	s.UsersSettings[userId] = userSetting

	_, err := s.userSettingsRepo.UpdateSongRequestsSettingsByID(ctx, userId, songReqSettings)
	if err != nil {
		log.Error("Error while saving song request settings:", err)
		return err
	}
	return nil

}

func (s *SettingsService) GetCommands(ctx context.Context, userId string) ([]model.CommonCommand, error) {
	cmdAndDescriptions, err := s.GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}
	return cmdAndDescriptions, nil
}

func (s *SettingsService) GetCommandsList() (*model.CommandList, error) {
	ctx, cancel := context.WithTimeout(s.ctxApp, 3*time.Second)
	defer cancel()

	var (
		commandsListResult model.CommandList
		descriptionResult  model.CommandsDescription
		errResult          error
	)
	commandsListCh := make(chan model.CommandList)
	descriptionCh := make(chan model.CommandsDescription)
	errCh := make(chan error)

	var wg sync.WaitGroup
	wg.Add(2)

	// get commandList
	go func() {
		defer wg.Done()
		commandsList, err := s.settingsOptionsRepo.GetCommandList(ctx)
		if err != nil {
			errCh <- err
			return
		}
		commandsListCh <- commandsList
	}()

	// get commandDescription
	go func() {
		defer wg.Done()
		commandsDescription, err := s.settingsOptionsRepo.GetCommandDescription(ctx)
		if err != nil {
			errCh <- err
			return
		}
		descriptionCh <- commandsDescription
	}()
	go func() {
		wg.Wait()
		close(commandsListCh)
		close(descriptionCh)
		close(errCh)
	}()

	for {
		select {
		case commandsList, ok := <-commandsListCh:
			if !ok {
				commandsListCh = nil
			} else {
				commandsListResult = commandsList
			}
		case descriptions, ok := <-descriptionCh:
			if !ok {
				descriptionCh = nil
			} else {
				descriptionResult = descriptions
			}
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
			} else {
				log.Error("Error:", err)
				errResult = err
			}
		}

		// Exit the loop when all channels are closed
		if commandsListCh == nil && descriptionCh == nil && errCh == nil {
			break
		}
	}
	if errResult != nil {
		return nil, errResult
	}
	addDescriptionToCommands(&commandsListResult, descriptionResult)
	return &commandsListResult, nil
}

func (s *SettingsService) AddDefaultSettingsForUser(ctx context.Context, user model.User) error {
	settings, err := s.userSettingsRepo.GetDefaultSettings(ctx)
	if err != nil {
		log.Error("Error while getting default settings:", err)
		return err
	}

	settings.Id = user.Id

	alias := settings.Aliases["!пл"]
	alias.Payload = alias.Payload + user.Channel
	settings.Aliases["!пл"] = alias

	// set default aliases to new user
	s.vkplService.ChannelsCommands.Channels[user.Id] = model.ChCommand{Aliases: settings.Aliases}
	s.UsersSettings[user.Id] = model.UserSettings{SongRequests: settings.SongRequests}

	// save to DB
	_, err = s.userSettingsRepo.InsertOne(ctx, settings)
	if err != nil {
		log.Error("Error whole copying default settings:", err)
		return err
	}

	return nil
}

func (s *SettingsService) AddCommandForUser(ctx context.Context, userId string, command *model.CommonCommand) ([]model.CommonCommand, error) {
	s.vkplService.ChannelsCommands.Channels[userId].Aliases[command.Alias] = model.CmdDetails{
		Command:     command.Command,
		AccessLevel: command.AccessLevel,
		Payload:     command.Payload,
	}

	_, err := s.userSettingsRepo.UpdateAliasesByID(ctx, userId, s.vkplService.ChannelsCommands.Channels[userId].Aliases)
	if err != nil {
		log.Error("Error while updating aliases:", err)
		return nil, err
	}

	cmds, err := s.GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func (s *SettingsService) EditCommandForUser(ctx context.Context, userId string, alias string, command *model.CommonCommand) ([]model.CommonCommand, error) {
	delete(s.vkplService.ChannelsCommands.Channels[userId].Aliases, alias)
	s.vkplService.ChannelsCommands.Channels[userId].Aliases[command.Alias] = model.CmdDetails{
		Command:     command.Command,
		AccessLevel: command.AccessLevel,
		Payload:     command.Payload,
	}
	_, err := s.userSettingsRepo.UpdateAliasesByID(ctx, userId, s.vkplService.ChannelsCommands.Channels[userId].Aliases)
	if err != nil {
		log.Error("Error while updating aliases:", err)
		return nil, err
	}
	cmds, err := s.GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func (s *SettingsService) DeleteCommandForUser(ctx context.Context, userId string, alias string) ([]model.CommonCommand, error) {
	delete(s.vkplService.ChannelsCommands.Channels[userId].Aliases, alias)
	//_, err := DB.GetCollection(DB.UserSettings).UpdateByID(ctx, userId, bson.M{"$set": bson.M{"aliases": vkplay.ChannelsCommands.Channels[userId].Aliases}})
	_, err := s.userSettingsRepo.UpdateAliasesByID(ctx, userId, s.vkplService.ChannelsCommands.Channels[userId].Aliases)
	if err != nil {
		log.Error("Error while delete aliases:", err)
		return nil, err
	}
	cmds, err := s.GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func (s *SettingsService) GetCommandsWithDescription(ctx context.Context, userId string) ([]model.CommonCommand, error) {
	var cmds []model.CommonCommand
	commandDescription, err := s.settingsOptionsRepo.GetCommandDescription(ctx)
	if err != nil {
		log.Error("Error while getting command description:", err)
		return nil, err
	}
	for item := range s.vkplService.ChannelsCommands.Channels[userId].Aliases {
		cmds = append(cmds, model.CommonCommand{
			Alias:       item,
			Command:     s.vkplService.ChannelsCommands.Channels[userId].Aliases[item].Command,
			Description: commandDescription.CommandsDescription[s.vkplService.ChannelsCommands.Channels[userId].Aliases[item].Command],
			AccessLevel: s.vkplService.ChannelsCommands.Channels[userId].Aliases[item].AccessLevel,
			Payload:     s.vkplService.ChannelsCommands.Channels[userId].Aliases[item].Payload,
		})
	}
	return cmds, nil
}

func (s *SettingsService) AddUsersSkipCommands(ctx context.Context, userId string) {
	votingCommands := map[string]bool{
		"SR_UsersSkipSongYes": true,
		"SR_UsersSkipSongNo":  true,
	}

	for _, alias := range s.vkplService.ChannelsCommands.Channels[userId].Aliases {
		if _, exists := votingCommands[alias.Command]; exists {
			votingCommands[alias.Command] = false
		}
	}

	if value, exists := votingCommands["SR_UsersSkipSongYes"]; exists && value {
		s.vkplService.ChannelsCommands.Channels[userId].Aliases["!фу"] = model.CmdDetails{
			Command:     "SR_UsersSkipSongYes",
			AccessLevel: 0,
			Payload:     "",
		}
	}

	if value, exists := votingCommands["SR_UsersSkipSongNo"]; exists && value {
		s.vkplService.ChannelsCommands.Channels[userId].Aliases["!деф"] = model.CmdDetails{
			Command:     "SR_UsersSkipSongNo",
			AccessLevel: 0,
			Payload:     "",
		}
	}

	_, err := s.userSettingsRepo.UpdateAliasesByID(ctx, userId, s.vkplService.ChannelsCommands.Channels[userId].Aliases)
	if err != nil {
		log.Error("Error while adding users skip commands:", err)
	}
}

func addDescriptionToCommands(cmdList *model.CommandList, descriptions model.CommandsDescription) {
	for cmd := range cmdList.Commands {
		for item := range cmdList.Commands[cmd].Items {
			cmdList.Commands[cmd].Items[item].Label = descriptions.CommandsDescription[cmdList.Commands[cmd].Items[item].Value]
		}
	}
}
