package chat

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service/settings"
	"HoBot_Backend/pkg/service/songRequest"
	"HoBot_Backend/pkg/service/youtube"
	"HoBot_Backend/pkg/socketio"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"

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
	addCommand("Greating_To_User", helloCommand)
	addCommand("SR_SongRequest", srAdd)
	addCommand("SR_SetVolume", srSetVolume)
	addCommand("SR_SkipSong", srSkip)
	addCommand("SR_DeleteSong", srDeleteSongRequest)
	addCommand("SR_PlayPause", srPlayPause)
	addCommand("SR_CurrentSong", srCurrentSong)
	addCommand("SR_MySong", srMySong)
	addCommand("SR_UsersSkipSongYes", srUsersSkipSongYes)
	addCommand("SR_UsersSkipSongNo", srUsersSkipSongNo)
	addCommand("Print_Text", printText)
	addCommand("Available_Commands", availableCommands)
}

func addCommand(name string, handler func(msg *ChatMsg, param string)) {
	Commands[name] = Command{
		Name:    name,
		Handler: handler,
	}
}

func helloCommand(msg *ChatMsg, param string) {
	txt := fmt.Sprintf("Hello, %s! https://live.vkplay.ru/hobot asdf https://google.com https://www.youtube.com/", msg.GetDisplayName())
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

	srSettings := settings.UsersSettings[msg.GetChannelId()].SongRequests

	if info.Views < srSettings.MinVideoViews {
		SendWhisperToUser(fmt.Sprintf("Слишком мало просмотров у видео. От %d просмотров", srSettings.MinVideoViews), msg.GetChannelId(), msg.GetUser())
		return
	}

	if srSettings.MaxDurationMinutes > 0 && info.Duration > srSettings.MaxDurationMinutes*60 {
		SendWhisperToUser(fmt.Sprintf("Слишком продолжительное видео. Максимальное время видео - %d минут(ы)",
			srSettings.MaxDurationMinutes), msg.GetChannelId(), msg.GetUser())
		return
	}

	count, err := songRequest.CountSongsByUser(msg.GetChannelId(), msg.GetDisplayName())
	if err != nil {
		return
	}

	if srSettings.MaxRequestsPerUser > 0 && count >= srSettings.MaxRequestsPerUser {
		SendWhisperToUser(
			fmt.Sprintf("Ваши заказы уже в плейлисте. Не больше %d заказов от пользователя на плейлист",
				settings.UsersSettings[msg.GetChannelId()].SongRequests.MaxRequestsPerUser),
			msg.GetChannelId(), msg.GetUser())
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

	id, err := songRequest.AddSongRequestToDB(sr)
	if err != nil {
		log.Error("Error while adding song request to db:", err)
		return
	}

	sr.Id = id

	socketio.Emit(msg.GetChannelId(), socketio.SongRequestAdded, sr)
	SendWhisperToUser("Реквест добавлен в очередь", msg.GetChannelId(), msg.GetUser())
}

func srSetVolume(msg *ChatMsg, param string) {
	var vol int
	switch {
	case param == "":
		v, err := settings.GetVolume(context.Background(), msg.GetChannelId())
		if err != nil {
			return
		}
		SendWhisperToUser(fmt.Sprintf("Текущая громкость: %v%%", v), msg.GetChannelId(), msg.GetUser())
		return
	case param[0] == '+' || param[0] == '-':
		value := param[1:]
		v, err := strconv.Atoi(value)
		if err != nil {
			return
		}
		if param[0] == '-' {
			v = -v
		}
		vol, err = settings.ChangeVolumeBy(msg.GetChannelId(), v)
		if err != nil {
			return
		}
	default:
		v, err := strconv.Atoi(param)
		if err != nil {
			return
		}
		vol = max(0, min(v, 100))

		err = settings.SaveVolume(context.Background(), msg.GetChannelId(), vol)
		if err != nil {
			return
		}
	}

	socketio.Emit(msg.GetChannelId(), socketio.SongRequestSetVolume, vol)
	SendWhisperToUser(fmt.Sprintf("Громкость реквестов установлена на %v%%", vol), msg.GetChannelId(), msg.GetUser())
}

func srSkip(msg *ChatMsg, param string) {
	err := songRequest.SkipSong(msg.GetChannelId())
	if err != nil {
		return
	}
	socketio.Emit(msg.GetChannelId(), socketio.SongRequestSkipSong, "")
	SendMessageToChannel(msg.GetDisplayName()+" пропустил реквест", msg.GetChannelId(), nil)
}

func srDeleteSongRequest(msg *ChatMsg, param string) {
	if param == "" {
		return
	}

	currentSong, err := songRequest.GetCurrentSong(msg.GetChannelId())
	if err != nil {
		return
	}

	if currentSong.YT_ID == param {
		SendWhisperToUser("Текущую песню можно только пропустить, для этого используйте команду пропуска", msg.GetChannelId(), msg.GetUser())
		return
	}

	song, err := songRequest.DeleteSongByYtId(msg.GetChannelId(), param)
	if err != nil {
		return
	}

	socketio.Emit(msg.GetChannelId(), socketio.SongRequestDeleteSong, song.Id)
	SendMessageToChannel(fmt.Sprintf("%s удалил песню \"%s\" от %s", msg.GetDisplayName(), song.Title, song.By), msg.GetChannelId(), nil)

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
		SendWhisperToUser("Сейчас ничего не играет", msg.GetChannelId(), msg.GetUser())
		return
	}

	SendWhisperToUser(fmt.Sprintf("Текущая песня: %s ( https://youtu.be/%s )", song.Title, song.YT_ID), msg.GetChannelId(), msg.GetUser())
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
				SendWhisperToUser("Твоя песня играет прямо сейчас!", msg.GetChannelId(), msg.GetUser())
				break
			}
			SendWhisperToUser(fmt.Sprintf("До твоей песни %v (~%s)", i, fmtDuration(t)), msg.GetChannelId(), msg.GetUser())
			break
		}
		timeToMySong += song.Length
	}
}

func srUsersSkipSongYes(msg *ChatMsg, param string) {
	log.Info("Triggered skip song by " + msg.GetDisplayName())
	if !settings.UsersSettings[msg.GetChannelId()].SongRequests.IsUsersSkipAllowed {
		return
	}

	if songRequest.VotesForSkip[msg.GetChannelId()] != nil {
		if !songRequest.VotesForSkip[msg.GetChannelId()].HasVoted(msg.GetUser().ID) {
			log.Infof("%s voted SKIP.(%d/%d)\n", msg.GetDisplayName(), songRequest.VotesForSkip[msg.GetChannelId()].GetCount()+1, settings.UsersSettings[msg.GetChannelId()].SongRequests.UsersSkipValue)
		} else {
			log.Infof("%s tryed to vote AGAIN. Rejected\n", msg.GetDisplayName())
			return
		}
	} else {
		log.Infof("%s voted SKIP.(%d/%d)\n", msg.GetDisplayName(), 1, settings.UsersSettings[msg.GetChannelId()].SongRequests.UsersSkipValue)
	}

	isSkipped := songRequest.VotesForSkipYes(msg.GetChannelId(), msg.GetUser().ID)

	if isSkipped {
		socketio.Emit(msg.GetChannelId(), socketio.SongRequestSkipSong, "")
		SendMessageToChannel("Зрители пропустили реквест", msg.GetChannelId(), nil)
		log.Info("Skipped song by users")
		return
	}

	if songRequest.VotesForSkip[msg.GetChannelId()].GetCount() == (settings.UsersSettings[msg.GetChannelId()].SongRequests.UsersSkipValue / 2) {
		SendMessageToChannel(fmt.Sprintf("Набрано голосов для пропуска песни: %d/%d", songRequest.VotesForSkip[msg.GetChannelId()].GetCount(), settings.UsersSettings[msg.GetChannelId()].SongRequests.UsersSkipValue), msg.GetChannelId(), nil)
	}
}

func srUsersSkipSongNo(msg *ChatMsg, param string) {
	log.Info("Triggered def song by " + msg.GetDisplayName())
	if !settings.UsersSettings[msg.GetChannelId()].SongRequests.IsUsersSkipAllowed {
		return
	}

	if songRequest.VotesForSkip[msg.GetChannelId()] != nil {
		if !songRequest.VotesForSkip[msg.GetChannelId()].HasVoted(msg.GetUser().ID) {
			log.Infof("%s voted DEF.(%d/%d)\n", msg.GetDisplayName(), songRequest.VotesForSkip[msg.GetChannelId()].GetCount()-1, settings.UsersSettings[msg.GetChannelId()].SongRequests.UsersSkipValue)
		}
	} else {
		log.Infof("%s voted DEF.(%d/%d)\n", msg.GetDisplayName(), -1, settings.UsersSettings[msg.GetChannelId()].SongRequests.UsersSkipValue)
	}

	songRequest.VotesForSkipNo(msg.GetChannelId(), msg.GetUser().ID)
}

func printText(msg *ChatMsg, param string) {
	SendMessageToChannel(param, msg.GetChannelId(), msg.GetUser())
}

func availableCommands(msg *ChatMsg, param string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var channelOwner model.User
	err := DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"_id": msg.GetChannelId()}).Decode(&channelOwner)
	if err != nil {
		return
	}

	/*	if !isBotModeratorAndSentMsg(msg, channelOwner) {
		return
	}*/

	commands, err := settings.GetCommandsWithDescription(ctx, msg.GetChannelId())
	if err != nil {
		return
	}

	accessLevel := 0
	if msg.GetUser().IsOwner {
		accessLevel = 2
	} else if msg.GetUser().IsChatModerator {
		accessLevel = 1
	}

	commandsSb := strings.Builder{}
	textCommandsSb := strings.Builder{}
	for _, v := range commands {
		if accessLevel < v.AccessLevel {
			continue
		}
		if v.Command != "Print_Text" {
			if commandsSb.Len() > 0 {
				commandsSb.WriteString(" | ")
			}
			commandsSb.WriteString(v.Alias + " - " + v.Description)
		} else {
			if textCommandsSb.Len() > 0 {
				textCommandsSb.WriteString(", ")
			}
			textCommandsSb.WriteString(v.Alias)
		}
	}

	resCommands := ""
	if commandsSb.Len() > 0 {
		resCommands += "Доступные Вам команды: " + commandsSb.String()
	}

	if textCommandsSb.Len() > 0 {
		if len(resCommands) > 0 {
			resCommands += " | "
		}
		resCommands += "Текстовые команды: " + textCommandsSb.String()
	}

	if accessLevel > 0 {
		resCommands += " | Помощь по командам - https://hobot.alwaysdata.net/p/help"
	}

	SendWhisperToUser(resCommands, msg.GetChannelId(), msg.GetUser())
}
