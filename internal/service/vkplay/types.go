package vkplay

type AuthResponse struct {
	ClientID     string `json:"clientId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

type AuthRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

//-----------

type Config struct {
	Id               string   `bson:"_id"`
	ChannelsAutoJoin []string `bson:"channelsAutoJoin"`
}

// for store commands
type ChannelCommands struct {
	Channels map[string]ChCommand `bson:"channels"`
}

type ChCommand struct {
	Aliases map[string]CmdDetails `bson:"aliases"`
}

type CmdDetails struct {
	Command     string `bson:"command"`
	AccessLevel int    `bson:"access_level"`
	Payload     string `bson:"payload"`
}

// -----------

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
			Counters struct {
				Subscribers int `json:"subscribers"`
			} `json:"counters"`
			CoverURL          string `json:"cover_url"`
			Description       string `json:"description"`
			Status            string `json:"status"`
			URL               string `json:"url"`
			WebSocketChannels struct {
				ChannelPoints        string `json:"channel_points"`
				Chat                 string `json:"chat"`
				Info                 string `json:"info"`
				LimitedChat          string `json:"limited_chat"`
				LimitedPrivateChat   string `json:"limited_private_chat"`
				PrivateChannelPoints string `json:"private_channel_points"`
				PrivateChat          string `json:"private_chat"`
				PrivateInfo          string `json:"private_info"`
			} `json:"web_socket_channels"`
		} `json:"channel"`
		Owner struct {
			AvatarURL            string `json:"avatar_url"`
			ExternalProfileLinks []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"external_profile_links"`
			ID                 int    `json:"id"`
			IsVerifiedStreamer bool   `json:"is_verified_streamer"`
			Nick               string `json:"nick"`
			NickColor          int    `json:"nick_color"`
		} `json:"owner"`
		Stream struct {
			Category struct {
				CoverURL string `json:"cover_url"`
				ID       string `json:"id"`
				Title    string `json:"title"`
				Type     string `json:"type"`
			} `json:"category"`
			Counters struct {
				Viewers int `json:"viewers"`
				Views   int `json:"views"`
			} `json:"counters"`
			EndedAt    int64  `json:"ended_at"`
			ID         string `json:"id"`
			PreviewURL string `json:"preview_url"`
			Reactions  []struct {
				Count int    `json:"count"`
				Type  string `json:"type"`
			} `json:"reactions"`
			SourceURLs []struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			} `json:"source_urls"`
			StartedAt int64  `json:"started_at"`
			Title     string `json:"title"`
			VideoID   string `json:"video_id"`
			VkVideo   struct {
				OwnerID string `json:"owner_id"`
				VideoID string `json:"video_id"`
			} `json:"vk_video"`
		} `json:"stream"`
	} `json:"data"`
}
