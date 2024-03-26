package vkplay

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	vkpl             Vkpl
	authVkpl         AuthResponse
	ctxParent        context.Context
	ChannelsCommands = ChannelCommands{
		Channels: make(map[string]ChCommand),
	}
)

func Start(ctx context.Context) {
	ctxParent = ctx
	err := connectWS()
	if err != nil {
		log.Error(err)
	}
	go listen()
	getCommandsFromDB()
}

func refreshVkplToken() error {
	log.Info("VKPL: Refreshing vkplay token")
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Jar: jar,
	}

	loginURL := "https://auth-ac.vkplay.ru/sign_in"
	loginData := url.Values{
		"login":                {os.Getenv("VKPL_LOGIN")},
		"password":             {os.Getenv("VKPL_PASSWORD")},
		"continue":             {"https://account.vkplay.ru/oauth2/?client_id=vkplay.live&response_type=code&skip_grants=1&state=%7B%22unregId%22%3A%22streams_web%3A75c4625e-0231-466c-a023-74db07d45ea0%22%2C%22from%22%3A%22%22%2C%22redirectAppId%22%3A%22streams_web%22%7D%2A%2A%2A-%2A%2A%2Avkplay&redirect_uri=https%3A%2F%2Flive.vkplay.ru%2Fapp%2Foauth_redirect"},
		"failure":              {"https://account.vkplay.ru/oauth2/login/?continue=https%3A%2F%2Faccount.vkplay.ru%2Foauth2%2Flogin%2F%3Fcontinue%3Dhttps%253A%252F%252Faccount.vkplay.ru%252Foauth2%252F%253Fclient_id%253Dlive.vkplay.ru%2526response_type%253Dcode%2526skip_grants%253D1%2526state%253D%25257B%252522unregId%252522%25253A%252522streams_web%25253A75c4625e-0231-466c-a023-74db07d45ea0%252522%25252C%252522from%252522%25253A%252522%252522%25252C%252522redirectAppId%252522%25253A%252522streams_web%252522%25257D%25252A%25252A%25252A-%25252A%25252A%25252Avkplay%2526redirect_uri%253Dhttps%25253A%25252F%25252Flive.vkplay.ru%25252Fapp%25252Foauth_redirect%26client_id%3Dlive.vkplay.ru%26lang%3Dru_RU&client_id=live.vkplay.ru&lang=ru_RU"},
		"g-recaptcha-response": {""},
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(loginData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://account.vkplay.ru")
	req.Header.Set("Referer", "https://account.vkplay.ru")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	urlVkpl, _ := url.Parse("https://live.vkplay.ru")
	cookies := jar.Cookies(urlVkpl)

	var authResponse AuthResponse
	var tmpClientID string
	for _, cookie := range cookies {
		fmt.Printf("cookie: %v\n", cookie)
		if cookie.Name == "_clientId" {
			tmpClientID = cookie.Value
			continue
		}
		if cookie.Name == "auth" {
			validCookie, _ := url.QueryUnescape(cookie.Value)
			err = json.Unmarshal([]byte(validCookie), &authResponse)
			if err != nil {
				log.Error("Error decoding auth cookie:", err)
				return fmt.Errorf("error decoding auth cookie: %w", err)
			}
		}
	}
	authResponse.ClientID = tmpClientID

	authVkpl = authResponse
	return nil
}

func saveVkplAuthToDB(auth AuthResponse) error {
	log.Info("VKPL: Saving vkplay auth to DB")
	ctx, cancel := context.WithTimeout(ctxParent, 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.Vkpl).ReplaceOne(ctx, bson.M{"_id": "auth"}, auth)
	if err != nil {
		log.Error("Error while inserting vkplay auth:", err)
		return err
	}
	return nil
}

func getVkplAuthFromDB() (AuthResponse, error) {
	log.Info("VKPL: Getting vkplay auth from DB")
	ctx, cancel := context.WithTimeout(ctxParent, 3*time.Second)
	defer cancel()
	var auth AuthResponse
	err := DB.GetCollection(DB.Vkpl).FindOne(ctx, bson.M{"_id": "auth"}).Decode(&auth)
	if err != nil {
		log.Error("Error while getting vkplay auth:", err)
		return AuthResponse{}, err
	}
	return auth, nil
}

func isAuthNeedRefresh() bool {
	if authVkpl.ExpiresAt < time.Now().Add(time.Minute*10).UnixMilli() {
		return true
	}
	return false
}

func GetVkplToken() string {
	if authVkpl == (AuthResponse{}) {
		authFromDB, err := getVkplAuthFromDB()
		if err != nil {
			log.Error("Error while getting vkplay auth from db:", err)
			return ""
		}
		authVkpl = authFromDB
	}

	if authVkpl == (AuthResponse{}) || isAuthNeedRefresh() {
		err := refreshVkplToken()
		if err != nil {
			log.Error("Error while refreshing vkplay token:", err)
			return ""
		}
		err = saveVkplAuthToDB(authVkpl)
		if err != nil {
			log.Error("Error while saving vkplay auth to db:", err)
			return ""
		}

	}
	return authVkpl.AccessToken
}

func getWsToken() string {
	authVkplToken := GetVkplToken()
	if authVkplToken == "" {
		return ""
	}

	req, err := http.NewRequest("GET", "https://api.live.vkplay.ru/v1/ws/connect", nil)
	if err != nil {
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+authVkplToken)
	req.Header.Add("X-From-Id", authVkpl.ClientID)
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
	json.Unmarshal(b, &token)
	return token["token"].(string)
}

func connectWS() error {
	vkpl.wsCounter = 0
	wsToken := getWsToken()
	if wsToken == "" {
		return fmt.Errorf("ws token is empty")
	}
	vkpl.wsToken = wsToken

	h := http.Header{
		"Origin": {"https://live.vkplay.ru"},
	}
	wsCon, resp, err := websocket.DefaultDialer.Dial("wss://pubsub.live.vkplay.ru/connection/websocket?cf_protocol_version=v2", h)
	if err != nil {
		log.Error("Error while connecting to ws: %d", resp.StatusCode)
		return err
	}

	vkpl.wsConnect = wsCon
	vkpl.wsCounter++
	t := fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":%d}`, vkpl.wsToken, vkpl.wsCounter)
	err = SendWSMessage([]byte(t))
	if err != nil {
		return err
	}

	_, _, err = vkpl.wsConnect.ReadMessage()
	if err != nil {
		log.Error("Error while reading ws message. Check:", err)
		return err
	}

	vkpl.wsCounter++
	err = joinAllChats()
	if err != nil {
		log.Error("Error while joining chat:", err)
		return err
	}
	return nil
}

func SendWSMessage(p []byte) error {
	vkpl.wsCounter++
	err := vkpl.wsConnect.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Error("Error while sending ws message:", err)
		return err
	}
	return nil
}

func ReadWSMessage() (p []byte, err error) {
	_, p, err = vkpl.wsConnect.ReadMessage()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func joinOrLeaveChat(channel string, join bool) error {
	vkpl.wsCounter++
	action := "subscribe"
	if !join {
		action = "unsubscribe"
	}
	p := fmt.Sprintf(`{"%s":{"channel":"channel-chat:%s"},"id":%d}`, action, channel, vkpl.wsCounter)
	err := SendWSMessage([]byte(p))
	if err != nil {
		return err
	}
	return nil
}

func joinAllChats() error {
	channels := getWsChannelsFromDB().ChannelsAutoJoin

	for _, channel := range channels {
		err := joinOrLeaveChat(channel, true)
		if err != nil {
			log.Error("Error while joining chat:", err)
			return err
		}
	}
	return nil
}

func listen() {
	for {
		p, err := ReadWSMessage()
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
			SendWSMessage([]byte("{}"))
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
				return
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
							return

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

func getWsChannelsFromDB() Config {
	var config Config
	ctx, cancel := context.WithTimeout(ctxParent, 3*time.Second)
	defer cancel()

	err := DB.GetCollection(DB.Config).FindOne(ctx, bson.M{"_id": "ws"}).Decode(&config)
	if err != nil {
		log.Error("Error while getting channels:", err)
		return Config{}
	}
	return config
}

func saveWsChannelsToDB(config Config) error {
	ctx, cancel := context.WithTimeout(ctxParent, 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.Config).UpdateByID(ctx, "ws", bson.M{"$set": config})
	if err != nil {
		log.Error("Error while saving channels:", err)
		return err
	}
	return nil
}

func AddUserToWs(userId string) error {
	// DB
	wsChannels := getWsChannelsFromDB()
	wsChannels.ChannelsAutoJoin = append(wsChannels.ChannelsAutoJoin, userId)
	err := saveWsChannelsToDB(wsChannels)
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

	return nil
}

func RemoveUserFromWs(userId string) error {
	// DB
	wsChannels := getWsChannelsFromDB()
	for i, v := range wsChannels.ChannelsAutoJoin {
		if v == userId {
			wsChannels.ChannelsAutoJoin = append(wsChannels.ChannelsAutoJoin[:i], wsChannels.ChannelsAutoJoin[i+1:]...)
			break
		}
	}
	err := saveWsChannelsToDB(wsChannels)
	if err != nil {
		log.Error("Error while removing user from ws:", err)
		return err
	}

	// WS
	err = joinOrLeaveChat(userId, false)
	if err != nil {
		log.Error("Error while leaving chat for removed user:", err)
		return err
	}
	return nil
}

func getCommandsFromDB() {
	var cmds ChannelCommands
	ctx, cancel := context.WithTimeout(ctxParent, 5*time.Second)
	defer cancel()

	// Set up the aggregation pipeline
	pipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", nil},
			{"channels", bson.D{{"$push", bson.D{
				{"k", "$_id"},
				{"v", bson.D{{"aliases", "$aliases"}}},
			}}}},
		}}},
		{{"$replaceRoot", bson.D{
			{"newRoot", bson.D{
				{"_id", "commands"},
				{"channels", bson.D{{"$arrayToObject", "$channels"}}},
			}},
		}}},
	}

	// Execute the aggregation
	cursor, err := DB.GetCollection(DB.UserSettings).Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Error while aggregating:", err)
	}
	defer cursor.Close(ctx)

	// Iterate over the result
	if cursor.Next(ctx) {
		err := cursor.Decode(&cmds)
		if err != nil {
			log.Error("Error while decoding:", err)
		}
	}

	ChannelsCommands = cmds
}

func SendMessageToChannel(msgText string, channel string, mention *User) {
	var msg []interface{}

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
	segments := re.Split(msgText, -1)
	matches := re.FindAllStringSubmatch(msgText, -1)

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
	req.Header.Add("Authorization", "Bearer "+GetVkplToken())
	req.Header.Add("X-From-Id", authVkpl.ClientID)

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

func CodeToToken(code string) (string, error) {
	reqUrl := "https://api.live.vkplay.ru/oauth/server/token"
	reqData := url.Values{
		"code":         {code},
		"grant_type":   {"authorization_code"},
		"redirect_uri": {os.Getenv("CLIENT_AUTH_REDIRECT")},
	}

	req, err := http.NewRequest("POST", reqUrl, bytes.NewBufferString(reqData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+os.Getenv("VKPL_APP_CREDEANTIALS"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		log.Error("Error while changing code to token")
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error while changing code to token, read body:", err)
		return "", err
	}

	var token map[string]interface{}
	json.Unmarshal(b, &token)
	return token["access_token"].(string), nil
}

func GetCurrentUserInfo(accessToken string) (model.CurrentUserVkpl, error) {
	req, err := http.NewRequest("GET", "https://apidev.live.vkplay.ru/v1/current_user", nil)
	if err != nil {
		return model.CurrentUserVkpl{}, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while getting ws token:", err)
		return model.CurrentUserVkpl{}, err
	}
	defer resp.Body.Close()

	var currentUserVkpl model.CurrentUserVkpl
	err = json.NewDecoder(resp.Body).Decode(&currentUserVkpl)
	if err != nil {
		return model.CurrentUserVkpl{}, err
	}

	return currentUserVkpl, nil
}

func InsertOrUpdateUser(ctx context.Context, user model.User) error {
	log.Info("VKPL: Inserting or updating user")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.Users).UpdateByID(ctx, user.Id, bson.M{"$set": user}, options.Update().SetUpsert(true))
	if err != nil {
		log.Error("Error while inserting or updating user:", err)
		return err
	}
	return nil
}
