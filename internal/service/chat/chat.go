package chat

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/service/vkplay"
	"HoBot_Backend/internal/service/voting"
	"HoBot_Backend/internal/socketio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	repoConfig "HoBot_Backend/internal/repository/config"
	repoUser "HoBot_Backend/internal/repository/user"
	repoVkpl "HoBot_Backend/internal/repository/vkpl"

	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
)

type ChatService struct {
	vkplWs   VkplWs
	wsConIds map[int]string

	ctxApp         context.Context
	vkplRepo       repoVkpl.Repository
	userRepo       repoUser.Repository
	configRepo     repoConfig.Repository
	socketioServer *socketio.SocketServer
	vkplService    *vkplay.VkplService
	votingService  *voting.VotingService
	commandService *CommandService
}

func NewChatService(ctxApp context.Context, socketioServer *socketio.SocketServer, vkplRepo repoVkpl.Repository, userRepo repoUser.Repository, configRepo repoConfig.Repository, vkplService *vkplay.VkplService, votingService *voting.VotingService) *ChatService {

	return &ChatService{
		vkplWs:         VkplWs{},
		wsConIds:       make(map[int]string),
		ctxApp:         ctxApp,
		vkplRepo:       vkplRepo,
		userRepo:       userRepo,
		configRepo:     configRepo,
		socketioServer: socketioServer,
		vkplService:    vkplService,
		votingService:  votingService,
		commandService: nil,
	}
}

func (s *ChatService) SetCommandService(commandService *CommandService) {
	s.commandService = commandService
}

func (s *ChatService) Start() {
	if s.commandService == nil {
		log.Error("Command service is not set")
		return
	}

	err := s.connectWS()
	if err != nil {
		log.Error(err)
	}

	go s.listen()
	go s.checkWsConns()
}
func (s *ChatService) listen() {
	for {
		p, err := s.readWSMessage()
		//log.Info(string(p))
		if err != nil {
			log.Error("Error while reading ws message:", err)
			log.Info("VKPL: Reconnecting to ws")
			err := s.connectWS()
			if err != nil {
				log.Error("Error while reconnecting to ws:", err)
			}
			continue
		}
		if isPING(p) {
			err := s.sendWSMessage([]byte("{}"))
			if err != nil {
				log.Error("Error while sending ws ping message:", err)
			}
		} else {
			var msg model.ChatMsg
			err = json.Unmarshal(p, &msg)
			if err != nil && !strings.Contains(string(p), "subscribe") {
				log.Error("Error while unmarshalling ws message:", err)
				// ---------- Block for printing error ----------
				dst := &bytes.Buffer{}
				if err := json.Indent(dst, p, "", "  "); err != nil {
					log.Error(err)
					//panic(err)
				}
				log.Error(dst.String())
				// ----------
				log.Info("VKPL: Invalid message:", string(p))
				continue
			}

			// --------------
			if strings.Contains(string(p), "subscribe") {
				//log.Info(string(p))
				lines := bytes.Split(p, []byte("\n"))
				for _, line := range lines {
					if len(bytes.TrimSpace(line)) == 0 {
						continue
					}

					if !json.Valid(line) {
						log.Info("Invalid JSON: %s", line)
						continue
					}

					type wsRes struct {
						ID int `json:"id"`
					}
					var wsMsg wsRes
					err = json.Unmarshal(line, &wsMsg)
					if err != nil {
						log.Error("WS2 Error while unmarshalling ws message:", err)
					}
					delete(s.wsConIds, wsMsg.ID)
				}

			}
			// --------------

			if msg.Push.Pub.Data.Type == "message" {
				var sb strings.Builder

				//================== Block for printing all data
				/*empJSON, err := json.MarshalIndent(msg.Push.Pub.Data.Data.Data, "", "  ")
				if err != nil {
					log.Fatalf(err.Error())
				}

				fmt.Printf("All Data: %s\n", string(empJSON))*/
				//==================

				for _, d := range msg.Push.Pub.Data.Data.Data {
					var content []interface{}

					if (d.Type == "text" || d.Type == "link") && d.Modificator == "" {
						err := json.Unmarshal([]byte(d.Content), &content)
						if err != nil {
							log.Error("Error while unmarshalling content:", err)
							// ----------
							dst := &bytes.Buffer{}
							if err := json.Indent(dst, p, "", "  "); err != nil {
								log.Error(err)
								//panic(err)
							}
							log.Error(dst.String())
							// ----------
							continue
						}
						sb.WriteString(content[0].(string))
					} else if d.Type == "smile" {
						sb.WriteString(d.Name)
					}
				}

				trimSb := strings.TrimSpace(sb.String())
				if len(trimSb) == 0 {
					continue
				}

				// Print each message
				//fmt.Printf("%s: %s\n", msg.GetDisplayName(), trimSb)

				// TEMP
				/*				if msg.IsSubscriber() {
								fmt.Println(msg.GetDisplayName())
							}*/

				alias, param := getAliasAndParamFromMessage(trimSb)

				if s.votingService.Voting[msg.GetChannelId(s.userRepo.GetUserIdByWs)] != nil && s.votingService.Voting[msg.GetChannelId(s.userRepo.GetUserIdByWs)].IsVotingInProgress {
					if value, isContained := s.votingService.Voting[msg.GetChannelId(s.userRepo.GetUserIdByWs)].VotingAnswers[alias]; isContained {
						if s.votingService.Voting[msg.GetChannelId(s.userRepo.GetUserIdByWs)].AddVote(msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser().ID, msg.GetUser().DisplayName, value) {
							s.socketioServer.Emit(msg.GetChannelId(s.userRepo.GetUserIdByWs), socketio.VotingVote, &voting.Vote{Name: msg.GetUser().DisplayName, Vote: value})
						}
						continue
					}
				}

				if !s.hasAccess(alias, &msg) {
					continue
				}

				cmd, payload := s.getCommandAndPayloadForAlias(alias, msg.GetChannelId(s.userRepo.GetUserIdByWs))
				if cmd != "" {
					if payload != "" {
						param = payload
					}
					if cmd == "Lasqa_KP" {
						param = trimSb
					}
					s.commandService.Commands[cmd].Handler(&msg, param)
				}
			}
		}
	}
}

func (s *ChatService) SendMessageToChannel(msgText string, channel string, mention *model.UserVk) {
	var msg []interface{}
	msgTextClear := msgText
	if len([]rune(msgTextClear)) > 495 {
		log.Info("Too long message: (", len([]rune(msgTextClear)), ") ", msgTextClear)
		msgTextClear = "Невозможно отобразить, слишком длинное сообщение."
	}

	ctx, cancel := context.WithTimeout(s.ctxApp, 5*time.Second)
	defer cancel()

	userDb, err := s.userRepo.GetUser(ctx, channel)
	if err != nil {
		log.Error("Error while getting user id:", err)
		return
	}

	// Adding mention if present
	if mention != nil {
		m := &MsgMentionContent{
			Type:        "mention",
			ID:          mention.ID,
			Nick:        mention.Nick,
			DisplayName: mention.DisplayName,
			Name:        mention.Name,
		}
		msg = append(msg, m)
	}

	// Parsing message text for links
	re := regexp.MustCompile(`(https?://[^\s]+)`)
	segments := re.Split(msgTextClear, -1)
	matches := re.FindAllStringSubmatch(msgTextClear, -1)

	// Adding segments and links to the message
	for i, seg := range segments {
		// Adding non-link segments
		if seg != "" {
			txt := &MsgTextContent{
				Modificator: "",
				Type:        "text",
				Content:     makeContentJSON(seg),
			}
			msg = append(msg, txt)
		}

		// Adding link blocks
		if i < len(matches) {
			match := matches[i][0]
			link := &MsgLinkContent{
				Type:    "link",
				Content: makeContentJSON(match),
				Url:     match,
			}
			msg = append(msg, link)
		}
	}

	// Adding block end
	txt := &MsgTextContent{
		Modificator: "BLOCK_END",
		Type:        "text",
		Content:     "",
	}
	msg = append(msg, txt)

	// Marshalling the message
	b, _ := json.Marshal(msg)

	body := strings.NewReader("data=" + string(b))

	// Creating and sending the request
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.live.vkvideo.ru/v1/blog/%s/public_video_stream/chat", userDb.Channel), body)
	if err != nil {
		log.Error("Error while sending message to channel:", err)
		return
	}

	req.Header.Add("Origin", "https://live.vkvideo.ru")
	req.Header.Add("Referer", fmt.Sprintf("https://live.vkvideo.ru/%s", channel))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+s.vkplService.GetVkplToken())
	req.Header.Add("X-From-Id", s.vkplService.AuthVkpl.ClientID)

	client := http.Client{}

	for attempt := 0; attempt < 3; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error while sending message to channel:", err)
			return
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errResp ErrorResponse
			if jerr := json.Unmarshal(body, &errResp); jerr != nil {
				log.Error("Failed to unmarshal error response:", jerr)
				break
			}

			shouldRetry := false
			for _, reason := range errResp.Data.Reasons {
				if reason.Type == "slow_mode_cooldown" {
					shouldRetry = true
					log.Info("Waiting for slow mode cooldown...")
					time.Sleep(time.Duration((attempt+1)*500) * time.Millisecond)
					break
				}
			}

			if shouldRetry {
				continue
			}

			// other errors
			log.Error("Error while sending message to channel:", string(body))
			break
		}

		break
	}

}

func (s *ChatService) SendWhisperToUser(msgText string, channel string, user *model.UserVk) {
	displayNameRunes := []rune(user.DisplayName)
	overhead := len(displayNameRunes) + 4 // 4 accounts for "/w " and a space

	const maxMessageLength = 495
	maxSegmentLength := maxMessageLength - overhead

	runes := []rune(msgText)

	if len(runes) <= maxSegmentLength {
		s.SendMessageToChannel("/w \""+user.DisplayName+"\" "+msgText, channel, nil)
		return
	}

	for i := 0; i < len(runes); i += maxSegmentLength {
		to := i + maxSegmentLength
		if to > len(runes) {
			to = len(runes)
		}
		segment := string(runes[i:to])
		s.SendMessageToChannel("/w \""+user.DisplayName+"\" "+segment, channel, nil)
	}
}

func (s *ChatService) AddUserToWs(user model.User) error {
	// DB
	wsChannels := s.configRepo.GetWsChannels(s.ctxApp)
	wsChannels.ChannelsAutoJoin = append(wsChannels.ChannelsAutoJoin, user.ChannelWS)
	err := s.configRepo.SaveWsChannels(s.ctxApp, wsChannels)
	if err != nil {
		log.Error("Error while adding user to ws:", err)
		return err
	}

	// WS
	err = s.joinOrLeaveChat(user.ChannelWS, true)
	if err != nil {
		log.Error("Error while joining chat for new user:", err)
		return err
	}

	s.SendMessageToChannel("Бот подключился к чату. Для нормальной работы боту необходимы права модератора. Выдать права боту можно командой \"/mod channel HoBOT\"", user.Id, nil)

	return nil
}

func (s *ChatService) RemoveUserFromWs(user model.User) error {
	// DB
	wsChannels := s.configRepo.GetWsChannels(s.ctxApp)
	for i, v := range wsChannels.ChannelsAutoJoin {
		if v == user.ChannelWS {
			wsChannels.ChannelsAutoJoin = append(wsChannels.ChannelsAutoJoin[:i], wsChannels.ChannelsAutoJoin[i+1:]...)
			break
		}
	}
	err := s.configRepo.SaveWsChannels(s.ctxApp, wsChannels)
	if err != nil {
		log.Error("Error while removing user from ws:", err)
		return err
	}

	s.SendMessageToChannel("Бот покинул чат.", user.Id, nil)

	// WS
	err = s.joinOrLeaveChat(user.ChannelWS, false)
	if err != nil {
		log.Error("Error while leaving chat for removed user:", err)
		return err
	}
	return nil
}

func (s *ChatService) UpdateUserInWs(oldUserId string, newUserId string) error {
	// DB
	wsChannels := s.configRepo.GetWsChannels(s.ctxApp)
	for i, v := range wsChannels.ChannelsAutoJoin {
		if v == oldUserId {
			wsChannels.ChannelsAutoJoin[i] = newUserId
			break
		}
	}
	err := s.configRepo.SaveWsChannels(s.ctxApp, wsChannels)
	if err != nil {
		log.Error("Error while updating user in ws:", err)
		return err
	}

	// WS
	err = s.joinOrLeaveChat(oldUserId, false)
	if err != nil {
		log.Error("Error while leaving chat for updated user:", err)
		return err
	}
	err = s.joinOrLeaveChat(newUserId, true)
	if err != nil {
		log.Error("Error while joining chat for updated user:", err)
		return err
	}
	return nil
}

func (s *ChatService) connectWS() error {
	s.vkplWs.wsCounter = 0
	wsToken := s.getWsToken()
	if wsToken == "" {
		return fmt.Errorf("ws token is empty")
	}
	s.vkplWs.wsToken = wsToken

	h := http.Header{
		"Origin": {"https://live.vkvideo.ru"},
	}
	wsCon, resp, err := websocket.DefaultDialer.Dial("wss://pubsub-dev.live.vkvideo.ru/connection/websocket?cf_protocol_version=v2", h)
	if err != nil {
		log.Error("Error while connecting to ws: %d", resp.StatusCode)
		return err
	}

	s.vkplWs.wsConnect = wsCon
	s.vkplWs.wsCounter++
	t := fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":%d}`, s.vkplWs.wsToken, s.vkplWs.wsCounter)
	err = s.sendWSMessage([]byte(t))
	if err != nil {
		return err
	}

	_, _, err = s.vkplWs.wsConnect.ReadMessage()
	if err != nil {
		log.Error("Error while reading ws message. Check:", err)
		return err
	}

	s.vkplWs.wsCounter++
	err = s.joinAllChats()
	if err != nil {
		log.Error("Error while joining chat:", err)
		return err
	}
	return nil
}

func (s *ChatService) getWsToken() string {
	authVkplToken := s.vkplService.GetVkplToken()
	if authVkplToken == "" {
		return ""
	}

	req, err := http.NewRequest("GET", "https://api.live.vkvideo.ru/v1/ws/connect", nil)
	if err != nil {
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+authVkplToken)
	req.Header.Add("X-From-Id", s.vkplService.AuthVkpl.ClientID)
	req.Header.Add("Origin", "https://live.vkvideo.ru")
	req.Header.Add("Referer", "https://live.vkvideo.ru/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while getting ws token:", err)
		return ""
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error while reading ws token:", err)
		return ""
	}

	var token map[string]interface{}
	err = json.Unmarshal(b, &token)
	if err != nil {
		log.Error("Error while unmarshaling ws token:", err)
		return ""
	}
	return token["token"].(string)
}

func (s *ChatService) sendWSMessage(p []byte) error {
	s.vkplWs.wsCounter++
	err := s.vkplWs.wsConnect.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Error("Error while sending ws message:", err)
		return err
	}
	return nil
}

func (s *ChatService) readWSMessage() (p []byte, err error) {
	_, p, err = s.vkplWs.wsConnect.ReadMessage()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ChatService) joinOrLeaveChat(channel string, join bool) error {
	s.vkplWs.wsCounter++
	action := "subscribe"
	if !join {
		action = "unsubscribe"
	}

	// ----------
	if join {
		s.wsConIds[s.vkplWs.wsCounter] = channel
	}
	// ----------

	p := fmt.Sprintf(`{"%s":{"channel":"channel-chat:%s"},"id":%d}`, action, channel, s.vkplWs.wsCounter)
	err := s.sendWSMessage([]byte(p))
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) joinAllChats() error {
	channels := s.configRepo.GetWsChannels(s.ctxApp).ChannelsAutoJoin

	for _, channel := range channels {
		//if (i+1)%20 == 0 {
		//time.Sleep(1 * time.Second)
		//}
		err := s.joinOrLeaveChat(channel, true)
		if err != nil {
			log.Error("Error while joining chat:", err)
			return err
		}
	}
	return nil
}

func (s *ChatService) isBotModeratorAndSentMsg(msg *model.ChatMsg, channelOwner model.User) bool {
	if !vkplay.IsBotHaveModeratorRights(channelOwner.Channel) {
		s.SendMessageToChannel("Для использования этой команды боту необходимы права модератора (для отправки личных сообщений)", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return false
	}
	return true
}

func (s *ChatService) getCommandAndPayloadForAlias(alias, channel string) (cmd, param string) {
	if chnl, ok := s.vkplService.ChannelsCommands.Channels[channel]; ok {
		cmd = chnl.Aliases[alias].Command
		if chnl.Aliases[alias].Payload != "" {
			param = chnl.Aliases[alias].Payload
		}
	}
	return
}

func (s *ChatService) hasAccess(alias string, msg *model.ChatMsg) bool {
	accessLevel := s.vkplService.ChannelsCommands.Channels[msg.GetChannelId(s.userRepo.GetUserIdByWs)].Aliases[alias].AccessLevel
	switch accessLevel {
	case 1:
		return msg.GetUser().IsChatModerator || msg.GetUser().IsChannelModerator || msg.GetUser().IsOwner
	case 2:
		return msg.GetUser().IsOwner
	default:
		return true
	}
}

func (s *ChatService) checkWsConns() {
	time.Sleep(10 * time.Second)
	log.Info("How many ws not connected:", len(s.wsConIds))
	log.Info("Channels:")
	for _, v := range s.wsConIds {
		log.Info(v)
	}

	//time.Sleep(5 * time.Second)
	//checkWsPresence()
}

/* func checkWsPresence() {
	//channels := vkplay.GetWsChannelsFromDB().ChannelsAutoJoin
	vkplWs.wsCounter++
	//p := fmt.Sprintf(`{"method": "presence", "params": {"channel":"channel-chat:%s"}}`, "18591758")
	p := fmt.Sprintf(`{"presence":{"channel":"channel-chat:%s"},"id":%d}`, "18591758", vkplWs.wsCounter)

	err := sendWSMessage([]byte(p))
	if err != nil {
		log.Error("Error while sending ws message:", err)
		return
	}
} */
