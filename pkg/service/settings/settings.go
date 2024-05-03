package settings

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service/vkplay"
	"context"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var UsersSettings = make(map[string]UserSettings)

func LoadSettings() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := DB.GetCollection(DB.UserSettings).Find(ctx, bson.M{})
	if err != nil {
		log.Error("Error while loading settings:", err)
		return
	}

	for cursor.Next(ctx) {
		var result UserSettings
		if err := cursor.Decode(&result); err != nil {
			log.Error("Error while decoding settings:", err)
			continue
		}
		result.Aliases = nil
		result.Volume = 0
		UsersSettings[result.Id] = result
	}
}

func SaveSongRequestSettings(ctx context.Context, userId string, songReqSettings SongRequestsSettings) error {
	userSetting := UsersSettings[userId]
	userSetting.SongRequests = songReqSettings
	UsersSettings[userId] = userSetting
	_, err := DB.GetCollection(DB.UserSettings).UpdateByID(ctx, userId, bson.M{"$set": bson.M{"song_requests": songReqSettings}})
	if err != nil {
		log.Error("Error while saving song request settings:", err)
		return err
	}
	return nil

}

func GetCommands(ctx context.Context, userId string) ([]model.CommonCommand, error) {
	cmdAndDescriptions, err := GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}
	return cmdAndDescriptions, nil
}

func GetCommandsList() (*model.CommandList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
		var commandsList model.CommandList
		err := DB.GetCollection(DB.SettingsOptions).FindOne(ctx, bson.M{"_id": "commandsList"}).Decode(&commandsList)
		if err != nil {
			errCh <- err
			return
		}
		commandsListCh <- commandsList
	}()

	// get commandDescription
	go func() {
		defer wg.Done()
		commandsDescription, err := getCommandDescription(ctx)
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

func getCommandDescription(ctx context.Context) (model.CommandsDescription, error) {
	var commandsDescription model.CommandsDescription
	err := DB.GetCollection(DB.SettingsOptions).FindOne(ctx, bson.M{"_id": "commandsDescription"}).Decode(&commandsDescription)
	if err != nil {
		log.Error("Error while getting command description:", err)
		return model.CommandsDescription{}, err
	}
	return commandsDescription, nil
}

func AddDefaultSettingsForUser(ctx context.Context, user model.User) error {
	var settings UserSettings
	err := DB.GetCollection(DB.UserSettings).FindOne(ctx, bson.M{"_id": "default"}).Decode(&settings)
	if err != nil {
		log.Error("Error while getting default settings:", err)
		return err
	}

	settings.Id = user.Id

	alias := settings.Aliases["!пл"]
	alias.Payload = alias.Payload + user.Channel
	settings.Aliases["!пл"] = alias

	// set default aliases to new user
	vkplay.ChannelsCommands.Channels[user.Id] = vkplay.ChCommand{Aliases: settings.Aliases}
	UsersSettings[user.Id] = UserSettings{SongRequests: settings.SongRequests}

	// save to DB
	_, err = DB.GetCollection(DB.UserSettings).InsertOne(ctx, settings)
	if err != nil {
		log.Error("Error whole copying default settings:", err)
		return err
	}

	return nil
}

func AddCommandForUser(ctx context.Context, userId string, command *model.CommonCommand) ([]model.CommonCommand, error) {
	vkplay.ChannelsCommands.Channels[userId].Aliases[command.Alias] = vkplay.CmdDetails{
		Command:     command.Command,
		AccessLevel: command.AccessLevel,
		Payload:     command.Payload,
	}

	_, err := DB.GetCollection(DB.UserSettings).UpdateByID(ctx, userId, bson.M{"$set": bson.M{"aliases": vkplay.ChannelsCommands.Channels[userId].Aliases}})
	if err != nil {
		log.Error("Error while updating aliases:", err)
		return nil, err
	}

	cmds, err := GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func EditCommandForUser(ctx context.Context, userId string, alias string, command *model.CommonCommand) ([]model.CommonCommand, error) {
	delete(vkplay.ChannelsCommands.Channels[userId].Aliases, alias)
	vkplay.ChannelsCommands.Channels[userId].Aliases[command.Alias] = vkplay.CmdDetails{
		Command:     command.Command,
		AccessLevel: command.AccessLevel,
		Payload:     command.Payload,
	}
	_, err := DB.GetCollection(DB.UserSettings).UpdateByID(ctx, userId, bson.M{"$set": bson.M{"aliases": vkplay.ChannelsCommands.Channels[userId].Aliases}})
	if err != nil {
		log.Error("Error while updating aliases:", err)
		return nil, err
	}
	cmds, err := GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func DeleteCommandForUser(ctx context.Context, userId string, alias string) ([]model.CommonCommand, error) {
	delete(vkplay.ChannelsCommands.Channels[userId].Aliases, alias)
	_, err := DB.GetCollection(DB.UserSettings).UpdateByID(ctx, userId, bson.M{"$set": bson.M{"aliases": vkplay.ChannelsCommands.Channels[userId].Aliases}})
	if err != nil {
		log.Error("Error while delete aliases:", err)
		return nil, err
	}
	cmds, err := GetCommandsWithDescription(ctx, userId)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func GetCommandsWithDescription(ctx context.Context, userId string) ([]model.CommonCommand, error) {
	var cmds []model.CommonCommand
	commandDescription, err := getCommandDescription(ctx)
	if err != nil {
		log.Error("Error while getting command description:", err)
		return nil, err
	}
	for item := range vkplay.ChannelsCommands.Channels[userId].Aliases {
		cmds = append(cmds, model.CommonCommand{
			Alias:       item,
			Command:     vkplay.ChannelsCommands.Channels[userId].Aliases[item].Command,
			Description: commandDescription.CommandsDescription[vkplay.ChannelsCommands.Channels[userId].Aliases[item].Command],
			AccessLevel: vkplay.ChannelsCommands.Channels[userId].Aliases[item].AccessLevel,
			Payload:     vkplay.ChannelsCommands.Channels[userId].Aliases[item].Payload,
		})
	}
	return cmds, nil
}

func addDescriptionToCommands(cmdList *model.CommandList, descriptions model.CommandsDescription) {
	for cmd := range cmdList.Commands {
		for item := range cmdList.Commands[cmd].Items {
			cmdList.Commands[cmd].Items[item].Label = descriptions.CommandsDescription[cmdList.Commands[cmd].Items[item].Value]
		}
	}
}

func SaveVolume(ctx context.Context, userId string, volume int) error {
	_, err := DB.GetCollection(DB.UserSettings).UpdateByID(ctx, userId, bson.M{"$set": bson.M{"volume": volume}})
	if err != nil {
		log.Error("Error while updating volume:", err)
		return err
	}
	return nil
}

func GetVolume(ctx context.Context, userId string) (int, error) {
	var userSettings UserSettings
	filter := bson.D{{"_id", userId}}
	opts := options.FindOne().SetProjection(bson.D{{"volume", 1}})
	err := DB.GetCollection(DB.UserSettings).FindOne(ctx, filter, opts).Decode(&userSettings)
	if err != nil {
		log.Error("Error while getting volume:", err)
		return 0, err
	}
	return userSettings.Volume, nil
}

func ChangeVolumeBy(userId string, volume int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vol, err := GetVolume(ctx, userId)
	if err != nil {
		return 0, err
	}

	vol += volume
	vol = max(0, min(vol, 100))

	err = SaveVolume(ctx, userId, vol)
	if err != nil {
		return 0, err
	}

	return vol, nil
}
