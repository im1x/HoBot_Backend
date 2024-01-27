package vkplay

import (
	"HoBot_Backend/pkg/service/youtube"
	"fmt"
)

type Command struct {
	Name    string
	Handler func(msg *ChatMsg, param string)
}

var Commands = make(map[string]Command)

func init() {
	AddCommand("Greating_To_User", helloCommand)
	AddCommand("SR_SongRequest", sr_songRequest)
}

func AddCommand(name string, handler func(msg *ChatMsg, param string)) {
	Commands[name] = Command{
		Name:    name,
		Handler: handler,
	}
}

func helloCommand(msg *ChatMsg, param string) {
	txt := fmt.Sprintf("Hello, %s!", msg.GetDisplayName())
	SendMessageToChannel(txt, msg.GetChannelName(), msg.GetUser())
}

func sr_songRequest(msg *ChatMsg, param string) {
	fmt.Println("sr_songRequest")
	if param == "" {
		return
	}
	fmt.Println(param)
	youtube.GetVideoInfo(param)
}
