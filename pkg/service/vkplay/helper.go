package vkplay

import (
	"fmt"
	"strings"
	"time"
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

func fmtDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h == 0 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
