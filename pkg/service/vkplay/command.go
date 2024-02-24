package vkplay

import (
	"HoBot_Backend/pkg/service/songRequest"
	"HoBot_Backend/pkg/service/youtube"
	"HoBot_Backend/pkg/socketio"
	"context"
	"strconv"

	//"HoBot_Backend/pkg/socketio"
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
	AddCommand("SR_SetVolume", srSetVolume)
	AddCommand("SR_SkipSong", srSkip)
	AddCommand("SR_PlayPause", srPlayPause)
	AddCommand("SR_CurrentSong", srCurrentSong)
	AddCommand("SR_MySong", srMySong)
}

func AddCommand(name string, handler func(msg *ChatMsg, param string)) {
	Commands[name] = Command{
		Name:    name,
		Handler: handler,
	}
}

func helloCommand(msg *ChatMsg, param string) {
	txt := fmt.Sprintf("Hello, %s! https://vkplay.live/hobot asdf https://google.com https://www.youtube.com/", msg.GetDisplayName())
	SendMessageToChannel(txt, msg.GetChannelId(), msg.GetUser())
}

func srAdd(msg *ChatMsg, param string) {
	if param == "" {
		return
	}

	if songRequest.IsPlaylistFull(msg.GetChannelId()) {
		SendMessageToChannel("Очередь заполнена", msg.GetChannelId(), msg.GetUser())
		return
	}
	vId := param
	if len(param) > 12 {
		id, err := youtube.ExtractVideoID(param)
		if err != nil {
			log.Error("Error while extracting video id:", err)
			return
		}
		vId = id
	}
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

	socketio.Emit(msg.GetChannelId(), socketio.SongRequestAdded, sr)
	SendMessageToChannel("Реквест добавлен в очередь", msg.GetChannelId(), msg.GetUser())
}

func srSetVolume(msg *ChatMsg, param string) {
	if param == "" {
		return
	}

	socketio.Emit(msg.GetChannelId(), socketio.SongRequestSetVolume, param)
	vol, err := strconv.Atoi(param)
	if err != nil {
		return
	}
	vol = max(0, min(vol, 100))

	SendMessageToChannel(fmt.Sprintf("Громкость реквестов установлена на %v%%", vol), msg.GetChannelId(), nil)
}

func srSkip(msg *ChatMsg, param string) {
	err := songRequest.SkipSong(msg.GetChannelId())
	if err != nil {
		return
	}
	socketio.Emit(msg.GetChannelId(), socketio.SongRequestSkipSong, "")
	SendMessageToChannel(msg.GetDisplayName()+" пропустил реквест", msg.GetChannelId(), nil)
}

func srPlayPause(msg *ChatMsg, param string) {
	socketio.Emit(msg.GetChannelId(), socketio.SongRequestPlayPause, "")
}

func srCurrentSong(msg *ChatMsg, param string) {
	song, err := songRequest.GetCurrentSong(msg.GetChannelId())
	if err != nil {
		return
	}

	if song.YT_ID == "" {
		SendMessageToChannel("Сейчас ничего не играет", msg.GetChannelId(), nil)
		return
	}

	SendMessageToChannel(fmt.Sprintf("Текущий реквест: %s (https://youtu.be/%s)", song.Title, song.YT_ID), msg.GetChannelId(), nil)
}

func srMySong(msg *ChatMsg, param string) {
	playlist, err := songRequest.GetPlaylist(context.Background(), msg.GetChannelId())
	if err != nil {
		return
	}

	timeToMySong := 0
	for i, song := range playlist {
		if song.By == msg.GetDisplayName() {
			t := time.Duration(timeToMySong) * time.Second
			if i == 0 {
				SendMessageToChannel("Твой реквест играет прямо сейчас!", msg.GetChannelId(), msg.GetUser())
				break
			}
			SendMessageToChannel(fmt.Sprintf("До твоего реквеста %v (~%s)", i, fmtDuration(t)), msg.GetChannelId(), msg.GetUser())
			break
		}
		timeToMySong += song.Length
	}
}
