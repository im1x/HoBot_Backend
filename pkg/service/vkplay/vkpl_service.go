package vkpl

import (
	DB "HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type Vkpl struct {
	wsConnect *websocket.Conn
	wsCounter int
	wsToken   string
}

var vkpl Vkpl

type AuthResponse struct {
	ClientID     string `json:"clientId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

var authVkpl AuthResponse

func refreshVkplToken() error {
	log.Info("VKPL: Refreshing vkpl token")
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
		"continue":             {"https://account.vkplay.ru/oauth2/?client_id=vkplay.live&response_type=code&skip_grants=1&state=%7B%22unregId%22%3A%22streams_web%3A75c4625e-0231-466c-a023-74db07d45ea0%22%2C%22from%22%3A%22%22%2C%22redirectAppId%22%3A%22streams_web%22%7D%2A%2A%2A-%2A%2A%2Avkplay&redirect_uri=https%3A%2F%2Fvkplay.live%2Fapp%2Foauth_redirect"},
		"failure":              {"https://account.vkplay.ru/oauth2/login/?continue=https%3A%2F%2Faccount.vkplay.ru%2Foauth2%2Flogin%2F%3Fcontinue%3Dhttps%253A%252F%252Faccount.vkplay.ru%252Foauth2%252F%253Fclient_id%253Dvkplay.live%2526response_type%253Dcode%2526skip_grants%253D1%2526state%253D%25257B%252522unregId%252522%25253A%252522streams_web%25253A75c4625e-0231-466c-a023-74db07d45ea0%252522%25252C%252522from%252522%25253A%252522%252522%25252C%252522redirectAppId%252522%25253A%252522streams_web%252522%25257D%25252A%25252A%25252A-%25252A%25252A%25252Avkplay%2526redirect_uri%253Dhttps%25253A%25252F%25252Fvkplay.live%25252Fapp%25252Foauth_redirect%26client_id%3Dvkplay.live%26lang%3Dru_RU&client_id=vkplay.live&lang=ru_RU"},
		"g-recaptcha-response": {""},
	}
	log.Info(loginData)
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

	urlVkpl, _ := url.Parse("https://vkplay.live")
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
				fmt.Println("Error decoding auth cookie:", err)
				return fmt.Errorf("error decoding auth cookie: %w", err)
			}
			//break
		}
	}
	authResponse.ClientID = tmpClientID
	log.Info(authResponse)
	defer resp.Body.Close()

	authVkpl = authResponse
	return nil
}

func saveVkplAuthToDB(auth AuthResponse) error {
	log.Info("VKPL: Saving vkpl auth")
	colVkpl := DB.GetCollection(DB.Vkpl)
	_, err := colVkpl.ReplaceOne(service.ctx, bson.M{"_id": "auth"}, auth)
	if err != nil {
		log.Error("Error while inserting vkpl auth:", err)
		return err
	}
	return nil
}

func getVkplAuthFromDB() (AuthResponse, error) {
	log.Info("VKPL: Getting vkpl auth from db")
	colVkpl := DB.GetCollection(DB.Vkpl)
	var auth AuthResponse
	err := colVkpl.FindOne(service.ctx, bson.M{"_id": "auth"}).Decode(&auth)
	if err != nil {
		log.Error("Error while getting vkpl auth:", err)
		return AuthResponse{}, err
	}
	return auth, nil
}

func isAuthNeedRefresh() bool {
	log.Info("VKPL: Checking if auth need refresh")
	if authVkpl.ExpiresAt < time.Now().Add(time.Minute*10).Unix() {
		return true
	}
	return false
}

func GetVkplToken() string {
	if authVkpl == (AuthResponse{}) {
		authFromDB, err := getVkplAuthFromDB()
		if err != nil {
			log.Error("Error while getting vkpl auth from db:", err)
			return ""
		}
		authVkpl = authFromDB
	}

	if authVkpl == (AuthResponse{}) || isAuthNeedRefresh() {
		err := refreshVkplToken()
		if err != nil {
			log.Error("Error while refreshing vkpl token:", err)
			return ""
		}
		err = saveVkplAuthToDB(authVkpl)
		if err != nil {
			log.Error("Error while saving vkpl auth to db:", err)
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

	req, err := http.NewRequest("GET", "https://api.vkplay.live/v1/ws/connect", nil)
	if err != nil {
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+authVkplToken)
	req.Header.Add("X-From-Id", authVkpl.ClientID)
	req.Header.Add("Origin", "https://vkplay.live")
	req.Header.Add("Referer", "https://vkplay.live/")

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

func (vkpl *Vkpl) ConnectWS() error {
	wsToken := getWsToken()
	if wsToken == "" {
		return fmt.Errorf("ws token is empty")
	}
	vkpl.wsToken = wsToken
	//vkpl.wsToken = "eyJhbGciOiJIUzI1NiJ9.eyJpbmZvIjp7Imp3dF9nZXRfaXAiOiIxMDkuMTEwLjc2LjE1MSJ9LCJzdWIiOiIxODM0Nzk4MSIsImV4cCI6MTcwNDI5NjI5Nn0.R345GHPrA7Cy3gXuedtSlxKQyGArvBN98oEvQKTE98M"
	log.Info("WS token: ", wsToken)

	h := http.Header{
		"Origin": {"https://vkplay.live"},
	}
	//wsCon, resp, err := websocket.DefaultDialer.Dial("wss://pubsub.vkplay.live/connection/websocket", h)
	wsCon, resp, err := websocket.DefaultDialer.Dial("wss://pubsub.vkplay.live/connection/websocket?cf_protocol_version=v2", h)
	if err != nil {
		log.Error("Error while connecting to ws: %d", resp.StatusCode)
		return err
	}

	vkpl.wsConnect = wsCon
	//t := fmt.Sprintf(`{"params": {"token": "%s", "name": "js"}, "id": %d}`, vkpl.wsToken, vkpl.wsCounter)
	vkpl.wsCounter++
	t := fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":%d}`, vkpl.wsToken, vkpl.wsCounter)
	err = vkpl.SendWSMessage([]byte(t))
	if err != nil {
		return err
	}

	respWsType, respWs, err := vkpl.wsConnect.ReadMessage()
	if err != nil {
		log.Error("Error while reading ws message. Check:", err)
		return err
	}

	log.Info(`VKPL-from-ws: respWsType %d respWs %s`, respWsType, string(respWs))

	vkpl.wsCounter++
	err = vkpl.joinAllChats()
	if err != nil {
		log.Error("Error while joining chat:", err)
		return err
	}
	vkpl.listen()

	return nil
}

func (vkpl *Vkpl) SendWSMessage(p []byte) error {
	vkpl.wsCounter++
	err := vkpl.wsConnect.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Error("Error while sending ws message:", err)
		return err
	}
	return nil
}

func (vkpl *Vkpl) ReadWSMessage() (p []byte, err error) {
	_, p, err = vkpl.wsConnect.ReadMessage()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (vkpl *Vkpl) joinChat(channel string) error {
	//p := fmt.Sprintf(`{"method": 1,"params":{"channel": "public-chat:%d"},"id":%d}`, "18591758", vkpl.wsCounter)
	vkpl.wsCounter++
	p := fmt.Sprintf(`{"subscribe":{"channel":"channel-chat:%s"},"id":%d}`, channel, vkpl.wsCounter)
	err := vkpl.SendWSMessage([]byte(p))
	if err != nil {
		return err
	}
	return nil
}

func (vkpl *Vkpl) joinAllChats() error {
	channels := []string{
		"18591758",
		"8845069",
	}

	for _, channel := range channels {
		err := vkpl.joinChat(channel)
		if err != nil {
			log.Error("Error while joining chat:", err)
			return err
		}
	}
	return nil
}

func (vkpl *Vkpl) listen() {
	go func() {
		for {
			p, err := vkpl.ReadWSMessage()
			if err != nil {
				log.Error("Error while reading ws message:", err)
				return
			}
			log.Info("VKPL-from-chat: ", string(p))
			if isPING(p) {
				vkpl.SendWSMessage([]byte("{}"))
			}
		}
	}()
}

func isPING(data []byte) bool {
	if len(data) != 2 { // Check the length first
		return false
	}
	return data[0] == '{' && data[1] == '}'
}
