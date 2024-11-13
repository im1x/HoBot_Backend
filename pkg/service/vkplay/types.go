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
