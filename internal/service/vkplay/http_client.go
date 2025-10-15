package vkplay

import (
	"HoBot_Backend/internal/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2/log"
)

func CodeToToken(code string) (string, error) {
	reqUrl := "https://api.live.vkvideo.ru/oauth/server/token"
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
	req, err := http.NewRequest("GET", "https://apidev.live.vkvideo.ru/v1/current_user", nil)
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

func IsBotHaveModeratorRights(chatName string) bool {
	userInfo, err := GetChatUserInfo(chatName, os.Getenv("BOT_VKPL_ID"))
	if err != nil {
		log.Error("Error while checking if bot have moderator rights:", err)
		return false
	}
	return userInfo.Data.User.IsModerator
}

func GetChatUserInfo(chatName string, userId string) (ChatUserDetails, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://apidev.live.vkvideo.ru/v1/chat/member?channel_url=%s&user_id=%s", chatName, userId), nil)
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

		if err == nil {
			err = errors.New("error while getting chat user details")
		}
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

func GetChannelInfo(channelName string) (ChannelInfo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://apidev.live.vkvideo.ru/v1/channel?channel_url=%s", channelName), nil)
	if err != nil {
		return ChannelInfo{}, err
	}

	req.Header.Set("Authorization", "Basic "+os.Getenv("VKPL_APP_CREDEANTIALS"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		log.Error("Error while getting channel info:", err)
		fmt.Println(resp.StatusCode)
		fmt.Println(err)
		fmt.Println(resp.Body)
		b, _ := io.ReadAll(resp.Body)
		fmt.Println(string(b))

		if err == nil {
			err = errors.New("error while getting channel info")
		}
		return ChannelInfo{}, err
	}
	defer resp.Body.Close()

	var channel ChannelInfo
	err = json.NewDecoder(resp.Body).Decode(&channel)
	if err != nil {
		log.Error("Error while decoding channel info:", err)
		return ChannelInfo{}, err
	}

	return channel, nil
}

func FollowUnfollowChannel(channelName string, isFollow bool, token string) error {
	urlReq := fmt.Sprintf("https://api.live.vkvideo.ru/v1/blog/%s/follow", channelName)
	if !isFollow {
		urlReq = fmt.Sprintf("https://api.live.vkvideo.ru/v1/blog/%s/unsubscribe", channelName)

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
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("host", "api.live.vkvideo.ru")
	req.Header.Add("Origin", "https://live.vkvideo.ru")
	req.Header.Add("Referer", fmt.Sprintf("https://live.vkvideo.ru/%s", channelName))

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

func GetUserIdWsByName(user string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.live.vkvideo.ru/v1/blog/%s/public_video_stream/chat/user/", user), nil)
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
