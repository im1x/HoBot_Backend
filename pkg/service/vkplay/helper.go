package vkplay

import (
	"strings"
)

func isPING(data []byte) bool {
	if len(data) != 2 { // Check the length first
		return false
	}
	return data[0] == '{' && data[1] == '}'
}

func getCommandFromMessage(message string) []string {
	if message == "" {
		return nil
	}
	//commandAndParam := strings.Fields(strings.ToLower(message))
	commandAndParam := strings.SplitN(strings.ToLower(message), " ", 3)
	if len(commandAndParam) > 1 {
		return commandAndParam[:2]
	} else {
		return commandAndParam[:1]
	}
}
func getCommandForAlias(command, channel string) (cmd string) {
	if chnl, ok := channelsCommands.Channels[channel]; ok {
		cmd = chnl.Aliases[command]
	}
	return
}
