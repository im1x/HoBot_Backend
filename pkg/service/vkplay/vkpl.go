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
	"net/url"
	"os"
	"strconv"
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

func FollowUnfollowChannel(channelName string, isFollow bool) error {
	urlReq := fmt.Sprintf("https://api.live.vkplay.ru/v1/blog/%s/follow", channelName)
	if !isFollow {
		urlReq = fmt.Sprintf("https://api.live.vkplay.ru/v1/blog/%s/unsubscribe", channelName)

	}
	req, err := http.NewRequest("POST", urlReq, nil)
	if err != nil {
		if isFollow {
			log.Error("Error while preparing following channel request:", err)
		} else {
			log.Error("Error while preparing unfollowing channel request:", err)
		}
		return err
	}

	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("Authorization", "Bearer "+GetVkplToken())
	req.Header.Add("host", "api.live.vkplay.ru")
	req.Header.Add("Origin", "https://live.vkplay.ru")
	req.Header.Add("Referer", fmt.Sprintf("https://live.vkplay.ru/%s", channelName))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if isFollow {
			log.Error("Error while following channel:", err)
		} else {
			log.Error("Error while unfollowing channel:", err)
		}
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response body: %v", err)
	}

	// Parse the response JSON
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error("FollowUnfollowChannel: Error unmarshalling response: %v", err, "isFollow ", isFollow)
	}

	// Check if response matches {"status": true}
	if status, ok := response["status"].(bool); ok && status {
		if isFollow {
			log.Info("Successfully followed channel:", channelName)
		} else {
			log.Info("Successfully unfollowed channel:", channelName)
		}
	} else {
		if isFollow {
			log.Error("Failed to follow channel:", channelName, "Response:", string(body))
		} else {
			log.Error("Failed to unfollow channel:", channelName, "Response:", string(body))
		}
	}

	return nil
}

/*func refreshVkplToken_Old() error {
	log.Info("VKPL: Refreshing vkplay token")
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Jar: jar,
	}

	// Init
	initURL := "https://auth-ac.vkplay.ru/api/v3/pub/auth/init"
	initJson := []byte(`{"csrfmiddlewaretoken":"","login":"` + os.Getenv("VKPL_LOGIN") + `","continue":"https://account.vkplay.ru/oauth2/?client_id=vkplay.live&response_type=code&skip_grants=1&state=%7B%22unregId%22%3A%22streams_web%3A83965df1-4f24-4c4d-a2d0-2b6c08b0f35f%22%2C%22from%22%3A%22%22%2C%22redirectAppId%22%3A%22streams_web%22%7D%2A%2A%2A-%2A%2A%2Avkplay&redirect_uri=https%3A%2F%2Fauth.live.vkplay.ru%2Fapp%2Foauth_redirect%3Ffrom%3Dvkplay_live","failure":"https://account.vkplay.ru/oauth2/login/?continue=https%3A%2F%2Faccount.vkplay.ru%2Foauth2%2Flogin%2F%3Fcontinue%3Dhttps%253A%252F%252Faccount.vkplay.ru%252Foauth2%252F%253Fclient_id%253Dvkplay.live%2526response_type%253Dcode%2526skip_grants%253D1%2526state%253D%25257B%252522unregId%252522%25253A%252522streams_web%25253A83965df1-4f24-4c4d-a2d0-2b6c08b0f35f%252522%25252C%252522from%252522%25253A%252522%252522%25252C%252522redirectAppId%252522%25253A%252522streams_web%252522%25257D%25252A%25252A%25252A-%25252A%25252A%25252Avkplay%2526redirect_uri%253Dhttps%25253A%25252F%25252Fauth.live.vkplay.ru%25252Fapp%25252Foauth_redirect%25253Ffrom%25253Dvkplay_live%26client_id%3Dvkplay.live%26lang%3Dru_RU&client_id=vkplay.live&lang=ru_RU"}`)
	req, err := http.NewRequest("POST", initURL, bytes.NewBuffer(initJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Origin", "https://account.vkplay.ru")
	req.Header.Set("Referer", "https://account.vkplay.ru")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	rd, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("INIT: Error while reading response body:", err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Error("INIT: Error while refreshing token:", err)
		log.Error("Status code:", resp.StatusCode)
		log.Error("Status:", resp.Status)
		log.Error("Response body:", string(rd))
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var initRes map[string]interface{}
	err = json.Unmarshal(rd, &initRes)
	if err != nil {
		return err
	}
	initToken := initRes["token"].(string)

	// Login
	loginURL := "https://auth-ac.vkplay.ru/api/v3/pub/auth/verify"
	loginData := map[string]string{
		"csrfmiddlewaretoken": "",
		"login":               os.Getenv("VKPL_LOGIN"),
		"password":            os.Getenv("VKPL_PASSWORD"),
		"token":               initToken,
	}

	loginJson, err := json.Marshal(loginData)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	req, err = http.NewRequest("POST", loginURL, bytes.NewBuffer(loginJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Origin", "https://account.vkplay.ru")
	req.Header.Set("Referer", "https://account.vkplay.ru")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}

	rd, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error("LOGIN: Error while reading response body:", err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Error("LOGIN: Error while refreshing token:", err)
		log.Error("Status code:", resp.StatusCode)
		log.Error("Status:", resp.Status)
		log.Error("Response body:", string(rd))
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var redirectRes map[string]interface{}
	err = json.Unmarshal(rd, &redirectRes)
	if err != nil {
		return err
	}
	redirectUrl := redirectRes["auth_redirect"].(string)
	redirectUrl = strings.ReplaceAll(redirectUrl, `\u0026`, "&")

	// Redirect
	req, err = http.NewRequest("GET", redirectUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Origin", "https://account.vkplay.ru")
	req.Header.Set("Referer", "https://account.vkplay.ru")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	rd, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error("REDIRECT: Error while reading response body:", err)
		return err
	}

	if resp.StatusCode >= 400 {
		log.Error("REDIRECT: Error while refreshing token:", err)
		log.Error("Status code:", resp.StatusCode)
		log.Error("Status:", resp.Status)
		log.Error("Response body:", string(rd))
		return err
	}

	defer resp.Body.Close()

	// Get cookies
	urlVkpl, _ := url.Parse("https://live.vkplay.ru")
	cookies := jar.Cookies(urlVkpl)

	var authResponse AuthResponse
	var tmpClientID string
	for _, cookie := range cookies {
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
}*/

func refreshVkplToken() error {
	log.Info("VKPL: Refreshing vkplay token!!!")

	reqUrl := "https://api.live.vkplay.ru/oauth/token/"
	reqData := url.Values{
		"response_type": {"code"},
		"refresh_token": {AuthVkpl.RefreshToken},
		"grant_type":    {"refresh_token"},
		"device_id":     {AuthVkpl.ClientID},
		"device_os":     {"streams_web"},
	}

	req, err := http.NewRequest("POST", reqUrl, bytes.NewBufferString(reqData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("origin", "https://live.vkplay.ru")
	req.Header.Set("referer", "https://live.vkplay.ru")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		log.Error("Error while refreshing token:", err)
		return err
	}
	defer resp.Body.Close()

	var refreshResponse AuthRefreshTokenResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Refresh token: Error while reading response body: ", err)
		return err
	}

	err = json.Unmarshal(body, &refreshResponse)
	if err != nil {
		log.Error("Error decoding refresh token:", err)
		return err
	}

	AuthVkpl.AccessToken = refreshResponse.AccessToken
	AuthVkpl.RefreshToken = refreshResponse.RefreshToken
	AuthVkpl.ExpiresAt = time.Now().Add(time.Second * time.Duration(refreshResponse.ExpiresIn)).UnixMilli()

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

func GetUserIdByName(user string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.live.vkplay.ru/v1/blog/%s/public_video_stream/chat/user/", user), nil)
	if err != nil {
		log.Error("Error while getting user id by name:", err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while getting user id by name:", err)
	}
	defer resp.Body.Close()

	type StreamerInfo struct {
		Data struct {
			Owner struct {
				ID int `json:"id"`
			} `json:"owner"`
		} `json:"data"`
	}

	var streamerInfo StreamerInfo
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error while reading user info:", err)
		return "", nil
	}

	err = json.Unmarshal(b, &streamerInfo)
	if err != nil {
		return "Error while unmarshal user info:", err
	}

	return strconv.Itoa(streamerInfo.Data.Owner.ID), nil

}
