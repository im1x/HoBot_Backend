package vkplay

import "strings"

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
	commandAndParam := strings.Split(message, " ")
	if len(commandAndParam) > 1 {
		return commandAndParam[:2]
	} else {
		return commandAndParam[:1]
	}
}
func isCommandAllowedOnChannel(command, channel string) string {
	if chnl, ok := channelsCommands.Channels[channel]; ok {
		if cmd, ok := chnl.Aliases[command]; ok {
			return cmd
		}
		return ""
	}
	return ""
}
