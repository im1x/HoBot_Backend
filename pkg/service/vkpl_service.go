package service

import (
	DB "HoBot_Backend/pkg/mongo"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type AuthResponse struct {
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
	for _, cookie := range cookies {
		fmt.Printf("cookie: %v\n", cookie)
		if cookie.Name == "auth" {
			validCookie, _ := url.QueryUnescape(cookie.Value)
			err = json.Unmarshal([]byte(validCookie), &authResponse)
			if err != nil {
				fmt.Println("Error decoding auth cookie:", err)
				return fmt.Errorf("error decoding auth cookie: %w", err)
			}
			break
		}
	}
	log.Info(authResponse)
	defer resp.Body.Close()

	authVkpl = authResponse
	return nil
}

func saveVkplAuthToDB(auth AuthResponse) error {
	log.Info("VKPL: Saving vkpl auth")
	colVkpl := DB.GetCollection(DB.Vkpl)
	_, err := colVkpl.ReplaceOne(ctx, bson.M{"_id": "auth"}, auth)
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
	err := colVkpl.FindOne(ctx, bson.M{"_id": "auth"}).Decode(&auth)
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
