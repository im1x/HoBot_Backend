package vkplay

import (
	"HoBot_Backend/pkg/service/songRequest"
	"HoBot_Backend/pkg/service/youtube"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"time"
)

type Command struct {
	Name    string
	Handler func(msg *ChatMsg, param string)
}

var Commands = make(map[string]Command)

func init() {
	AddCommand("Greating_To_User", helloCommand)
	AddCommand("SR_SongRequest", srAdd)
}

func AddCommand(name string, handler func(msg *ChatMsg, param string)) {
	Commands[name] = Command{
		Name:    name,
		Handler: handler,
	}
}

func helloCommand(msg *ChatMsg, param string) {
	txt := fmt.Sprintf("Hello, %s!", msg.GetDisplayName())
	SendMessageToChannel(txt, msg.GetChannelId(), msg.GetUser())
}

func srAdd(msg *ChatMsg, param string) {
	fmt.Println("srAdd")
	fmt.Println("param:", param)
	if param == "" {
		return
	}

	vId := param
	if len(param) > 12 {
		fmt.Println("param > 12")
		id, err := youtube.ExtractVideoID(param)
		if err != nil {
			log.Error("Error while extracting video id:", err)
			return
		}
		fmt.Println("id:", id)
		vId = id
	}
	fmt.Println("Ready to get info")
	info, err := youtube.GetVideoInfo(vId)
	if err != nil {
		return
	}

	var sr = songRequest.SongRequest{
		ChannelId: msg.GetChannelId(),
		By:        msg.GetDisplayName(),
		Requested: time.Now().Format(time.RFC3339),
		YT_ID:     vId,
		Title:     info.Title,
		Length:    info.Duration,
		Views:     info.Views,
		Start:     0,
		End:       0,
	}

	err = songRequest.AddSongRequestToDB(sr)
	if err != nil {
		log.Error("Error while adding song request to db:", err)
		return
	}

	SendMessageToChannel("Song request added to queue", msg.GetChannelId(), msg.GetUser())
}
