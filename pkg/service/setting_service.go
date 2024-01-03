package service

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
)

func GetCommandsList() (*model.CommandList, error) {
	var (
		commandsListResult model.CommandList
		aliasesResult      model.CommandsAliases
		errResult          error
	)
	commandsListCh := make(chan model.CommandList)
	aliasesCh := make(chan model.CommandsAliases)
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

	// get commandsAliases
	go func() {
		defer wg.Done()
		var commandsAliases model.CommandsAliases
		err := DB.GetCollection(DB.SettingsOptions).FindOne(ctx, bson.M{"_id": "commandsAliases"}).Decode(&commandsAliases)
		if err != nil {
			errCh <- err
			return
		}
		aliasesCh <- commandsAliases
	}()
	go func() {
		wg.Wait()
		close(commandsListCh)
		close(aliasesCh)
		close(errCh)
	}()

	for {
		select {
		case commandsList, ok := <-commandsListCh:
			if !ok {
				commandsListCh = nil
			} else {
				fmt.Println("Received commandsList:", commandsList)
				commandsListResult = commandsList
			}
		case aliases, ok := <-aliasesCh:
			if !ok {
				aliasesCh = nil
			} else {
				fmt.Println("Received aliases:", aliases)
				aliasesResult = aliases
			}
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
			} else {
				fmt.Println("Error:", err)
				errResult = err
			}
		}

		// Exit the loop when all channels are closed
		if commandsListCh == nil && aliasesCh == nil && errCh == nil {
			break
		}
	}
	fmt.Println("All Queries Completed")
	fmt.Println("Result:", commandsListResult, aliasesResult, errResult)
	if errResult != nil {
		return nil, errResult
	}
	addDescriptionToCommands(&commandsListResult, aliasesResult)
	return &commandsListResult, nil
	/*var commandsList model.CommandList
	err := DB.GetCollection(DB.SettingsOptions).FindOne(ctx, bson.M{"_id": "commandsList"}).Decode(&commandsList)
	if err != nil {
		return nil, err
	}
	return &commandsList, nil*/
}

/*func AddCommand(userId string, command *model.NewCommand) (model.CommandList, error) {
	if chnl, ok := channelsCommands.Channels[channel]; ok {

	}
	return model.CommandList{}, nil
}*/

func addDescriptionToCommands(cmdList *model.CommandList, aliases model.CommandsAliases) {
	for cmd := range cmdList.Commands {
		for item := range cmdList.Commands[cmd].Items {
			cmdList.Commands[cmd].Items[item].Label = aliases.CommandsAliases[cmdList.Commands[cmd].Items[item].Value]
		}
	}
}
