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
	commandAndParam := strings.Fields(strings.ReplaceAll(message, "\u00a0", " "))
	if len(commandAndParam) < 2 {
		commandAndParam = append(commandAndParam, "")
	}
	return commandAndParam[:2]
}
func getCommandForAlias(alias, channel string) (cmd string) {
	if chnl, ok := ChannelsCommands.Channels[channel]; ok {
		cmd = chnl.Aliases[alias].Command
	}
	return
}
