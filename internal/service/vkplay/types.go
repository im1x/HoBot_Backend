package vkplay

type AuthRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

//-----------

type ChatUserDetails struct {
	Data struct {
		User struct {
			ID           int    `json:"id"`
			Nick         string `json:"nick"`
			IsOwner      bool   `json:"is_owner"`
			IsModerator  bool   `json:"is_moderator"`
			RegisteredAt int    `json:"registered_at"`
		} `json:"user"`
		Statistics struct {
			ChatMessagesCount  int `json:"chat_messages_count"`
			PermanentBansCount int `json:"permanent_bans_count"`
			TemporaryBansCount int `json:"temporary_bans_count"`
			TotalWatchedTime   int `json:"total_watched_time"`
		} `json:"statistics"`
	} `json:"data"`
}

type ChannelInfo struct {
	Data struct {
		Channel struct {
			URL               string `json:"url"`
			CoverURL          string `json:"cover_url"`
			WebSocketChannels struct {
				Chat                 string `json:"chat"`
				PrivateChat          string `json:"private_chat"`
				Info                 string `json:"info"`
				PrivateInfo          string `json:"private_info"`
				ChannelPoints        string `json:"channel_points"`
				PrivateChannelPoints string `json:"private_channel_points"`
				LimitedChat          string `json:"limited_chat"`
				LimitedPrivateChat   string `json:"limited_private_chat"`
			} `json:"web_socket_channels"`
			Status   string `json:"status"`
			Counters struct {
				Subscribers int `json:"subscribers"`
			} `json:"counters"`
			Description string `json:"description"`
			ID          int    `json:"id"`
			AvatarURL   string `json:"avatar_url"`
			Nick        string `json:"nick"`
			NickColor   int    `json:"nick_color"`
		} `json:"channel"`
		Owner struct {
			AvatarURL            string        `json:"avatar_url"`
			Nick                 string        `json:"nick"`
			NickColor            int           `json:"nick_color"`
			ID                   int           `json:"id"`
			IsVerifiedStreamer   bool          `json:"is_verified_streamer"`
			ExternalProfileLinks []interface{} `json:"external_profile_links"`
		} `json:"owner"`
		Stream struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			StartedAt int    `json:"started_at"`
			EndedAt   int    `json:"ended_at"`
			Counters  struct {
				Viewers int `json:"viewers"`
				Views   int `json:"views"`
			} `json:"counters"`
			Reactions []interface{} `json:"reactions"`
			Category  struct {
				Title    string `json:"title"`
				Type     string `json:"type"`
				ID       string `json:"id"`
				CoverURL string `json:"cover_url"`
			} `json:"category"`
			PreviewURL string        `json:"preview_url"`
			SourceUrls []interface{} `json:"source_urls"`
			VideoID    string        `json:"video_id"`
			VkVideo    interface{}   `json:"vk_video"`
			Slot       struct {
				ID  int    `json:"id"`
				URL string `json:"url"`
			} `json:"slot"`
			Status string `json:"status"`
		} `json:"stream"`
		Streams []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			StartedAt int    `json:"started_at"`
			EndedAt   int    `json:"ended_at"`
			Counters  struct {
				Viewers int `json:"viewers"`
				Views   int `json:"views"`
			} `json:"counters"`
			Reactions []interface{} `json:"reactions"`
			Category  struct {
				Title    string `json:"title"`
				Type     string `json:"type"`
				ID       string `json:"id"`
				CoverURL string `json:"cover_url"`
			} `json:"category"`
			PreviewURL string        `json:"preview_url"`
			SourceUrls []interface{} `json:"source_urls"`
			VideoID    string        `json:"video_id"`
			VkVideo    interface{}   `json:"vk_video"`
			Slot       struct {
				ID  int    `json:"id"`
				URL string `json:"url"`
			} `json:"slot"`
			Status string `json:"status"`
		} `json:"streams"`
	} `json:"data"`
}
