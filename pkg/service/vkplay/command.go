package vkplay

import (
	"fmt"
)

type Command struct {
	Name        string
	Handler     func(msg *ChatMsg)
	Permissions int
}

var Commands = make(map[string]Command)

func init() {
	AddCommand("Greating_To_User", helloCommand, 0)
}

func AddCommand(name string, handler func(msg *ChatMsg), permissions int) {
	Commands[name] = Command{
		Name:        name,
		Handler:     handler,
		Permissions: permissions,
	}
}

func helloCommand(msg *ChatMsg) {
	txt := fmt.Sprintf("Hello, %s!", msg.GetDisplayName())
	SendMessageToChannel(txt, msg.GetChannelName(), msg.GetUser())
}
