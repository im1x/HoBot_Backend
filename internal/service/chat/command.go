package chat

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/service/settings"
	"HoBot_Backend/internal/service/songRequest"
	"HoBot_Backend/internal/service/youtube"
	"HoBot_Backend/internal/socketio"
	"context"
	"strconv"
	"strings"

	"fmt"
	"time"

	repoSongRequests "HoBot_Backend/internal/repository/songrequests"
	repoStatistics "HoBot_Backend/internal/repository/statistics"
	repoUser "HoBot_Backend/internal/repository/user"
	repoUserSettings "HoBot_Backend/internal/repository/usersettings"

	"github.com/gofiber/fiber/v2/log"
)

type Command struct {
	Name    string
	Handler func(msg *model.ChatMsg, param string)
}

//var Commands = make(map[string]Command)

type CommandService struct {
	appCtx             context.Context
	Commands           map[string]Command
	userRepo           repoUser.Repository
	songRequestRepo    repoSongRequests.Repository
	statisticsRepo     repoStatistics.Repository
	userSettingsRepo   repoUserSettings.Repository
	settingsService    *settings.SettingsService
	songRequestService *songRequest.SongRequestService
	socketioServer     *socketio.SocketServer
	chatService        *ChatService
	lasqaService       *LasqaService
}

func NewCommandService(appCtx context.Context, userRepo repoUser.Repository, songRequestRepo repoSongRequests.Repository, statisticsRepo repoStatistics.Repository, userSettingsRepo repoUserSettings.Repository, settingsService *settings.SettingsService, songRequestService *songRequest.SongRequestService, socketioServer *socketio.SocketServer, chatService *ChatService, lasqaService *LasqaService) *CommandService {

	commandService := &CommandService{
		appCtx:             appCtx,
		Commands:           make(map[string]Command),
		userRepo:           userRepo,
		songRequestRepo:    songRequestRepo,
		statisticsRepo:     statisticsRepo,
		userSettingsRepo:   userSettingsRepo,
		settingsService:    settingsService,
		songRequestService: songRequestService,
		socketioServer:     socketioServer,
		chatService:        chatService,
		lasqaService:       lasqaService,
	}

	commandService.init()
	return commandService
}

func (s *CommandService) init() {
	s.addCommand("Greating_To_User", s.helloCommand)
	s.addCommand("SR_SongRequest", s.srAdd)
	s.addCommand("SR_SetVolume", s.srSetVolume)
	s.addCommand("SR_SkipSong", s.srSkip)
	s.addCommand("SR_DeleteSong", s.srDeleteSongRequest)
	s.addCommand("SR_PlayPause", s.srPlayPause)
	s.addCommand("SR_CurrentSong", s.srCurrentSong)
	s.addCommand("SR_MySong", s.srMySong)
	s.addCommand("SR_UsersSkipSongYes", s.srUsersSkipSongYes)
	s.addCommand("SR_UsersSkipSongNo", s.srUsersSkipSongNo)
	s.addCommand("Print_Text", s.printText)
	s.addCommand("Available_Commands", s.availableCommands)
	// Lasqa
	s.addCommand("Lasqa_KP", s.lasqaService.LasqaKp)
}

func (s *CommandService) addCommand(name string, handler func(msg *model.ChatMsg, param string)) {
	s.Commands[name] = Command{
		Name:    name,
		Handler: handler,
	}
}

func (s *CommandService) helloCommand(msg *model.ChatMsg, param string) {
	txt := fmt.Sprintf("Hello, %s! https://live.vkvideo.ru/hobot asdf https://google.com https://www.youtube.com/", msg.GetDisplayName())
	//SendMessageToChannel(txt, msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())

	paramInt, _ := strconv.Atoi(param)
	s.chatService.SendWhisperToUser(txt, msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
	time.Sleep(time.Duration(paramInt) * time.Millisecond)
	s.chatService.SendWhisperToUser(txt, msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
}

func (s *CommandService) srAdd(msg *model.ChatMsg, param string) {
	if param == "" {
		return
	}

	if s.songRequestRepo.IsPlaylistFull(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs)) {
		s.chatService.SendMessageToChannel("Очередь заполнена", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
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

	srSettings := s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests

	if info.Views < srSettings.MinVideoViews {
		s.chatService.SendWhisperToUser(fmt.Sprintf("Слишком мало просмотров у видео. От %d просмотров", srSettings.MinVideoViews), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	if srSettings.MaxDurationMinutes > 0 && info.Duration > srSettings.MaxDurationMinutes*60 {
		s.chatService.SendWhisperToUser(fmt.Sprintf("Слишком продолжительное видео. Максимальное время видео - %d минут(ы)",
			srSettings.MaxDurationMinutes), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	count, err := s.songRequestRepo.CountSongsByUser(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetDisplayName())
	if err != nil {
		return
	}

	if srSettings.MaxRequestsPerUser > 0 && count >= srSettings.MaxRequestsPerUser {
		s.chatService.SendWhisperToUser(
			fmt.Sprintf("Ваши заказы уже в плейлисте. Не больше %d заказов от пользователя на плейлист",
				s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.MaxRequestsPerUser),
			msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	sr := model.SongRequest{
		ChannelId: msg.GetChannelId(s.userRepo.GetUserIdByWs),
		By:        msg.GetDisplayName(),
		Requested: time.Now().Format(time.RFC3339),
		YT_ID:     vId,
		Title:     info.Title,
		Length:    info.Duration,
		Views:     info.Views,
		Start:     0,
		End:       0,
	}

	id, err := s.songRequestRepo.AddSongRequestToDB(s.appCtx, sr)
	if err != nil {
		log.Error("Error while adding song request to db:", err)
		return
	}
	sr.Id = id

	s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.SongRequestAdded, sr)
	s.chatService.SendWhisperToUser("Реквест добавлен в очередь", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
	s.statisticsRepo.IncField(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), repoStatistics.SongRequestsAdd)
}

func (s *CommandService) srSetVolume(msg *model.ChatMsg, param string) {
	var vol int
	switch {
	case param == "":
		v, err := s.userSettingsRepo.GetVolume(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs))
		if err != nil {
			return
		}
		s.chatService.SendWhisperToUser(fmt.Sprintf("Текущая громкость: %v%%", v), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
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
		vol, err = s.userSettingsRepo.ChangeVolumeBy(msg.GetChannelId(s.userRepo.GetUserIdByWs), v)
		if err != nil {
			return
		}
	default:
		v, err := strconv.Atoi(param)
		if err != nil {
			return
		}
		vol = max(0, min(v, 100))

		err = s.userSettingsRepo.SaveVolume(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), vol)
		if err != nil {
			return
		}
	}

	s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.SongRequestSetVolume, vol)
	s.chatService.SendWhisperToUser(fmt.Sprintf("Громкость реквестов установлена на %v%%", vol), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
}

func (s *CommandService) srSkip(msg *model.ChatMsg, param string) {
	err := s.songRequestService.SkipSong(msg.GetChannelId(s.userRepo.GetUserIdByWs))
	if err != nil {
		return
	}
	s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.SongRequestSkipSong, "")
	s.chatService.SendMessageToChannel(msg.GetDisplayName()+" пропустил реквест", msg.GetChannelId(s.userRepo.GetUserIdByWs), nil)
	s.statisticsRepo.IncField(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), repoStatistics.SongRequestsSkipByCommand)
}

func (s *CommandService) srDeleteSongRequest(msg *model.ChatMsg, param string) {
	if param == "" {
		return
	}

	currentSong, err := s.songRequestRepo.GetCurrentSong(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs))
	if err != nil {
		return
	}

	if currentSong.YT_ID == param {
		s.chatService.SendWhisperToUser("Текущую песню можно только пропустить, для этого используйте команду пропуска", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	song, err := s.songRequestRepo.DeleteSongByYtId(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), param)
	if err != nil {
		return
	}

	s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.SongRequestDeleteSong, song.Id)
	s.chatService.SendMessageToChannel(fmt.Sprintf("%s удалил песню \"%s\" от %s", msg.GetDisplayName(), song.Title, song.By), msg.GetChannelId(s.userRepo.GetUserIdByWs), nil)

}

func (s *CommandService) srPlayPause(msg *model.ChatMsg, param string) {
	s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.SongRequestPlayPause, "")
}

func (s *CommandService) srCurrentSong(msg *model.ChatMsg, param string) {
	song, err := s.songRequestRepo.GetCurrentSong(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs))
	if err != nil {
		return
	}

	if song.YT_ID == "" {
		s.chatService.SendWhisperToUser("Сейчас ничего не играет", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	s.chatService.SendWhisperToUser(fmt.Sprintf("Текущая песня: %s ( https://youtu.be/%s )", song.Title, song.YT_ID), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
}

func (s *CommandService) srMySong(msg *model.ChatMsg, param string) {
	playlist, err := s.songRequestRepo.GetPlaylist(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs))
	if err != nil {
		return
	}

	timeToMySong := 0
	for i, song := range playlist {
		if song.By == msg.GetDisplayName() {
			t := time.Duration(timeToMySong) * time.Second
			if i == 0 {
				s.chatService.SendWhisperToUser("Твоя песня играет прямо сейчас!", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
				break
			}
			s.chatService.SendWhisperToUser(fmt.Sprintf("До твоей песни %v (~%s)", i, fmtDuration(t)), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
			break
		}
		timeToMySong += song.Length
	}
}

func (s *CommandService) srUsersSkipSongYes(msg *model.ChatMsg, param string) {
	log.Info("Triggered skip song by " + msg.GetDisplayName())
	if !s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.IsUsersSkipAllowed {
		return
	}

	if s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)] != nil {
		if !s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)].HasVoted(msg.GetUser().ID) {
			log.Infof("%s voted SKIP.(%d/%d)\n", msg.GetDisplayName(), s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)].GetCount()+1, s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.UsersSkipValue)
		} else {
			log.Infof("%s tryed to vote AGAIN. Rejected\n", msg.GetDisplayName())
			return
		}
	} else {
		log.Infof("%s voted SKIP.(%d/%d)\n", msg.GetDisplayName(), 1, s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.UsersSkipValue)
	}

	isSkipped := s.songRequestService.VotesForSkipYes(msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser().ID)

	if isSkipped {
		s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.SongRequestSkipSong, "")
		s.chatService.SendMessageToChannel("Зрители пропустили реквест", msg.GetChannelId(s.userRepo.GetUserIdByWs), nil)
		log.Info("Skipped song by users")
		s.statisticsRepo.IncField(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), repoStatistics.SongRequestsSkipByUsers)
		return
	}

	if s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)].GetCount() == (s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.UsersSkipValue / 2) {
		s.chatService.SendMessageToChannel(fmt.Sprintf("Набрано голосов для пропуска песни: %d/%d", s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)].GetCount(), s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.UsersSkipValue), msg.GetChannelId(s.userRepo.GetUserIdByWs), nil)
	}
}

func (s *CommandService) srUsersSkipSongNo(msg *model.ChatMsg, param string) {
	log.Info("Triggered def song by " + msg.GetDisplayName())
	if !s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.IsUsersSkipAllowed {
		return
	}

	if s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)] != nil {
		if !s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)].HasVoted(msg.GetUser().ID) {
			log.Infof("%s voted DEF.(%d/%d)\n", msg.GetDisplayName(), s.songRequestService.VotesForSkip[msg.GetChannelId(s.userRepo.GetUserIdByWs)].GetCount()-1, s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.UsersSkipValue)
		}
	} else {
		log.Infof("%s voted DEF.(%d/%d)\n", msg.GetDisplayName(), -1, s.settingsService.UsersSettings[msg.GetChannelId(s.userRepo.GetUserIdByWs)].SongRequests.UsersSkipValue)
	}

	s.songRequestService.VotesForSkipNo(msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser().ID)
}

func (s *CommandService) printText(msg *model.ChatMsg, param string) {
	s.chatService.SendMessageToChannel(param, msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
	s.statisticsRepo.IncField(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs), repoStatistics.PrintTextByCommand)
}

func (s *CommandService) availableCommands(msg *model.ChatMsg, param string) {
	/*channelOwner, err := s.userRepo.GetUser(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs))
	if err != nil {
		return
	}

		if !isBotModeratorAndSentMsg(msg, channelOwner) {
		return
	}*/

	commands, err := s.settingsService.GetCommandsWithDescription(s.appCtx, msg.GetChannelId(s.userRepo.GetUserIdByWs))
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

	s.chatService.SendWhisperToUser(resCommands, msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
}
