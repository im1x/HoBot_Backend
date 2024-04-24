package chat

import (
	"HoBot_Backend/pkg/service/vkplay"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var vkplWs VkplWs

func Start() {
	err := connectWS()
	if err != nil {
		log.Error(err)
	}

	go listen()
}
func listen() {
	for {
		p, err := readWSMessage()
		if err != nil {
			log.Error("Error while reading ws message:", err)
			log.Info("VKPL: Reconnecting to ws")
			err := connectWS()
			if err != nil {
				log.Error("Error while reconnecting to ws:", err)
			}
			continue
		}
		if isPING(p) {
			err := sendWSMessage([]byte("{}"))
			if err != nil {
				log.Error("Error while sending ws ping message:", err)
			}
		} else {
			var msg ChatMsg
			err = json.Unmarshal(p, &msg)
			if err != nil {
				log.Error("Error while unmarshalling ws message:", err)

				// ---------- Block for printing error ----------
				dst := &bytes.Buffer{}
				if err := json.Indent(dst, p, "", "  "); err != nil {
					log.Error(err)
					//panic(err)
				}
				log.Error(dst.String())
				// ----------
				continue
			}

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
					}
				}

				trimSb := strings.TrimSpace(sb.String())
				if len(trimSb) == 0 {
					continue
				}

				// Print each message
				//fmt.Printf("%s: %s\n", msg.GetDisplayName(), trimSb)

				alias, param := getAliasAndParamFromMessage(trimSb)
				if !hasAccess(alias, &msg) {
					continue
				}

				cmd, payload := getCommandAndPayloadForAlias(alias, msg.GetChannelId())
				if cmd != "" {
					if payload != "" {
						param = payload
					}
					Commands[cmd].Handler(&msg, param)
				}
			}
		}
	}
}

func SendMessageToChannel(msgText string, channel string, mention *User) {
	var msg []interface{}
	msgTextClear := prepareStringForSend(msgText)
	if len(msgTextClear) > 495 {
		msgTextClear = "Невозможно отобразить, слишком длинное сообщение."
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
				Content:     fmt.Sprintf("[\"%s \",\"unstyled\",[]]", seg),
			}
			msg = append(msg, txt)
		}

		// Adding link blocks
		if i < len(matches) {
			match := matches[i][0]
			link := &MsgLinkContent{
				Type:    "link",
				Content: fmt.Sprintf("[\"%s \",\"unstyled\",[]]", match),
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
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.live.vkplay.ru/v1/blog/%s/public_video_stream/chat", channel), body)
	if err != nil {
		log.Error("Error while sending message to channel:", err)
		return
	}

	req.Header.Add("Origin", "https://live.vkplay.ru")
	req.Header.Add("Referer", fmt.Sprintf("https://live.vkplay.ru/%s", channel))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+vkplay.GetVkplToken())
	req.Header.Add("X-From-Id", vkplay.AuthVkpl.ClientID)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while sending message to channel:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		rd, _ := io.ReadAll(resp.Body)
		log.Error(string(rd))
	}
}

func SendWhisperToUser(msgText string, channel string, user *User) {
	SendMessageToChannel("/w "+user.DisplayName+" "+msgText, channel, nil)
}

func AddUserToWs(userId string) error {
	// DB
	wsChannels := vkplay.GetWsChannelsFromDB()
	wsChannels.ChannelsAutoJoin = append(wsChannels.ChannelsAutoJoin, userId)
	err := vkplay.SaveWsChannelsToDB(wsChannels)
	if err != nil {
		log.Error("Error while adding user to ws:", err)
		return err
	}

	// WS
	err = joinOrLeaveChat(userId, true)
	if err != nil {
		log.Error("Error while joining chat for new user:", err)
		return err
	}

	SendMessageToChannel("Бот подключился к чату. Для нормальной работы боту необходимы права модератора.", userId, nil)

	return nil
}

func RemoveUserFromWs(userId string) error {
	// DB
	wsChannels := vkplay.GetWsChannelsFromDB()
	for i, v := range wsChannels.ChannelsAutoJoin {
		if v == userId {
			wsChannels.ChannelsAutoJoin = append(wsChannels.ChannelsAutoJoin[:i], wsChannels.ChannelsAutoJoin[i+1:]...)
			break
		}
	}
	err := vkplay.SaveWsChannelsToDB(wsChannels)
	if err != nil {
		log.Error("Error while removing user from ws:", err)
		return err
	}

	SendMessageToChannel("Бот покинул чат.", userId, nil)

	// WS
	err = joinOrLeaveChat(userId, false)
	if err != nil {
		log.Error("Error while leaving chat for removed user:", err)
		return err
	}
	return nil
}

func connectWS() error {
	vkplWs.wsCounter = 0
	wsToken := getWsToken()
	if wsToken == "" {
		return fmt.Errorf("ws token is empty")
	}
	vkplWs.wsToken = wsToken

	h := http.Header{
		"Origin": {"https://live.vkplay.ru"},
	}
	wsCon, resp, err := websocket.DefaultDialer.Dial("wss://pubsub.live.vkplay.ru/connection/websocket?cf_protocol_version=v2", h)
	if err != nil {
		log.Error("Error while connecting to ws: %d", resp.StatusCode)
		return err
	}

	vkplWs.wsConnect = wsCon
	vkplWs.wsCounter++
	t := fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":%d}`, vkplWs.wsToken, vkplWs.wsCounter)
	err = sendWSMessage([]byte(t))
	if err != nil {
		return err
	}

	_, _, err = vkplWs.wsConnect.ReadMessage()
	if err != nil {
		log.Error("Error while reading ws message. Check:", err)
		return err
	}

	vkplWs.wsCounter++
	err = joinAllChats()
	if err != nil {
		log.Error("Error while joining chat:", err)
		return err
	}
	return nil
}

func getWsToken() string {
	authVkplToken := vkplay.GetVkplToken()
	if authVkplToken == "" {
		return ""
	}

	req, err := http.NewRequest("GET", "https://api.live.vkplay.ru/v1/ws/connect", nil)
	if err != nil {
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+authVkplToken)
	req.Header.Add("X-From-Id", vkplay.AuthVkpl.ClientID)
	req.Header.Add("Origin", "https://live.vkplay.ru")
	req.Header.Add("Referer", "https://live.vkplay.ru/")

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

func sendWSMessage(p []byte) error {
	vkplWs.wsCounter++
	err := vkplWs.wsConnect.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Error("Error while sending ws message:", err)
		return err
	}
	return nil
}

func readWSMessage() (p []byte, err error) {
	_, p, err = vkplWs.wsConnect.ReadMessage()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func joinOrLeaveChat(channel string, join bool) error {
	vkplWs.wsCounter++
	action := "subscribe"
	if !join {
		action = "unsubscribe"
	}
	p := fmt.Sprintf(`{"%s":{"channel":"channel-chat:%s"},"id":%d}`, action, channel, vkplWs.wsCounter)
	err := sendWSMessage([]byte(p))
	if err != nil {
		return err
	}
	return nil
}

func joinAllChats() error {
	channels := vkplay.GetWsChannelsFromDB().ChannelsAutoJoin

	for _, channel := range channels {
		err := joinOrLeaveChat(channel, true)
		if err != nil {
			log.Error("Error while joining chat:", err)
			return err
		}
	}
	return nil
}
