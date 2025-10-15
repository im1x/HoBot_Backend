package vkplay

import (
	"HoBot_Backend/internal/model"
	repoUserSettings "HoBot_Backend/internal/repository/usersettings"
	repoVkpl "HoBot_Backend/internal/repository/vkpl"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type VkplService struct {
	ctxApp           context.Context
	AuthVkpl         model.AuthResponse
	ChannelsCommands model.ChannelCommands
	userSettingsRepo repoUserSettings.Repository
	vkplRepo         repoVkpl.Repository
}

func NewVkplService(ctx context.Context, userSettingsRepo repoUserSettings.Repository, vkplRepo repoVkpl.Repository) *VkplService {
	serviceData := &VkplService{
		ctxApp:   ctx,
		AuthVkpl: model.AuthResponse{},
		ChannelsCommands: model.ChannelCommands{
			Channels: make(map[string]model.ChCommand),
		},
		userSettingsRepo: userSettingsRepo,
		vkplRepo:         vkplRepo,
	}
	serviceData.ChannelsCommands = userSettingsRepo.GetCommands(ctx)

	return serviceData
}

func (s *VkplService) GetVkplToken() string {
	if s.AuthVkpl == (model.AuthResponse{}) {
		authFromDB, err := s.vkplRepo.GetAuth(s.ctxApp)
		if err != nil {
			log.Error("Error while getting vkplay auth from db:", err)
			return ""
		}
		s.AuthVkpl = authFromDB
	}

	if s.AuthVkpl == (model.AuthResponse{}) || s.isAuthNeedRefresh() {
		err := s.refreshVkplToken()
		if err != nil {
			log.Error("Error while refreshing vkplay token:", err)
			return ""
		}
		err = s.vkplRepo.SaveAuth(s.ctxApp, s.AuthVkpl)
		if err != nil {
			log.Error("Error while saving vkplay auth to db:", err)
			return ""
		}

	}
	return s.AuthVkpl.AccessToken
}

func (s *VkplService) refreshVkplToken() error {
	log.Info("VKPL: Refreshing vkplay token!!!")

	reqUrl := "https://api.live.vkvideo.ru/oauth/token/"
	reqData := url.Values{
		"response_type": {"code"},
		"refresh_token": {s.AuthVkpl.RefreshToken},
		"grant_type":    {"refresh_token"},
		"device_id":     {s.AuthVkpl.ClientID},
		"device_os":     {"streams_web"},
	}

	req, err := http.NewRequest("POST", reqUrl, bytes.NewBufferString(reqData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("origin", "https://live.vkvideo.ru")
	req.Header.Set("referer", "https://live.vkvideo.ru")
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

	s.AuthVkpl.AccessToken = refreshResponse.AccessToken
	s.AuthVkpl.RefreshToken = refreshResponse.RefreshToken
	s.AuthVkpl.ExpiresAt = time.Now().Add(time.Second * time.Duration(refreshResponse.ExpiresIn)).UnixMilli()

	return nil
}

func (s *VkplService) isAuthNeedRefresh() bool {
	return s.AuthVkpl.ExpiresAt < time.Now().Add(time.Minute*10).UnixMilli()
}
