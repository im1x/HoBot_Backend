package vkplay

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

var (
	AuthVkpl         AuthResponse
	ChannelsCommands = ChannelCommands{
		Channels: make(map[string]ChCommand),
	}
	ctxParent context.Context
)

func Start(ctx context.Context) {
	ctxParent = ctx
	getCommandsFromDB()
}

func GetVkplToken() string {
	if AuthVkpl == (AuthResponse{}) {
		authFromDB, err := getVkplAuthFromDB()
		if err != nil {
			log.Error("Error while getting vkplay auth from db:", err)
			return ""
		}
		AuthVkpl = authFromDB
	}

	if AuthVkpl == (AuthResponse{}) || isAuthNeedRefresh() {
		err := refreshVkplToken()
		if err != nil {
			log.Error("Error while refreshing vkplay token:", err)
			return ""
		}
		err = saveVkplAuthToDB(AuthVkpl)
		if err != nil {
			log.Error("Error while saving vkplay auth to db:", err)
			return ""
		}

	}
	return AuthVkpl.AccessToken
}

func GetWsChannelsFromDB() Config {
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

func SaveWsChannelsToDB(config Config) error {
	ctx, cancel := context.WithTimeout(ctxParent, 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.Config).UpdateByID(ctx, "ws", bson.M{"$set": config})
	if err != nil {
		log.Error("Error while saving channels:", err)
		return err
	}
	return nil
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
	err = json.Unmarshal(b, &token)
	if err != nil {
		return "", err
	}
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

func GetChatUserInfo(chatName string, userId string) (ChatUserDetails, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://apidev.live.vkplay.ru/v1/chat/member?channel_url=%s&user_id=%s", chatName, userId), nil)
	if err != nil {
		return ChatUserDetails{}, err
	}

	req.Header.Set("Authorization", "Basic "+os.Getenv("VKPL_APP_CREDEANTIALS"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		log.Error("Error while getting chat user details:", err)
		fmt.Println(resp.StatusCode)
		fmt.Println(err)
		fmt.Println(resp.Body)
		b, _ := io.ReadAll(resp.Body)
		fmt.Println(string(b))

		return ChatUserDetails{}, err
	}
	defer resp.Body.Close()

	var user ChatUserDetails
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Error("Error while decoding chat user details:", err)
		return ChatUserDetails{}, err
	}

	return user, nil
}

func IsBotHaveModeratorRights(chatName string) bool {
	userInfo, err := GetChatUserInfo(chatName, os.Getenv("BOT_VKPL_ID"))
	if err != nil {
		log.Error("Error while checking if bot have moderator rights:", err)
		return false
	}
	return userInfo.Data.User.IsModerator
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

	AuthVkpl = authResponse
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
	return AuthVkpl.ExpiresAt < time.Now().Add(time.Minute*10).UnixMilli()
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
