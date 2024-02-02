package vkplay

import (
	"github.com/gorilla/websocket"
	"strings"
)

type Vkpl struct {
	wsConnect *websocket.Conn
	wsCounter int
	wsToken   string
}

type AuthResponse struct {
	ClientID     string `json:"clientId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// -----------
type User struct {
	Roles []struct {
		ID        string `json:"id"`
		Priority  int    `json:"priority"`
		SmallURL  string `json:"smallUrl"`
		LargeURL  string `json:"largeUrl"`
		Name      string `json:"name"`
		MediumURL string `json:"mediumUrl"`
	} `json:"roles"`
	ID                 int    `json:"id"`
	IsOwner            bool   `json:"isOwner"`
	DisplayName        string `json:"displayName"`
	CreatedAt          int64  `json:"createdAt"`
	Nick               string `json:"nick"`
	AvatarURL          string `json:"avatarUrl"`
	IsChannelModerator bool   `json:"isChannelModerator"`
	Name               string `json:"name"`
	IsChatModerator    bool   `json:"isChatModerator"`
	HasAvatar          bool   `json:"hasAvatar"`
	VkplayProfileLink  string `json:"vkplayProfileLink"`
	IsVerifiedStreamer bool   `json:"isVerifiedStreamer"`
	NickColor          int    `json:"nickColor"`
	Badges             []struct {
		LargeURL    string `json:"largeUrl"`
		MediumURL   string `json:"mediumUrl"`
		IsCreated   bool   `json:"isCreated"`
		Name        string `json:"name"`
		SmallURL    string `json:"smallUrl"`
		Achievement struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"achievement"`
		ID string `json:"id"`
	} `json:"badges"`
}

type ChatMsg struct {
	Push struct {
		Channel string `json:"channel"`
		Pub     struct {
			Data struct {
				Data struct {
					Styles []any `json:"styles"`
					User   User  `json:"user"`
					Parent struct {
						Styles    []any `json:"styles"`
						ID        int   `json:"id"`
						Author    User  `json:"author"`
						IsPrivate bool  `json:"isPrivate"`
						CreatedAt int64 `json:"createdAt"`
						Data      []struct {
							Modificator string `json:"modificator"`
							Type        string `json:"type"`
							Content     string `json:"content"`
						} `json:"data"`
					} `json:"parent"`
					Data []struct {
						Name        string `json:"name,omitempty"`
						DisplayName string `json:"displayName,omitempty"`
						Nick        string `json:"nick,omitempty"`
						NickColor   string `json:"nickColor,omitempty"`
						BlogURL     string `json:"blogUrl,omitempty"`
						SmallURL    string `json:"smallUrl,omitempty"`
						MediumURL   string `json:"mediumUrl,omitempty"`
						LargeURL    string `json:"largeUrl,omitempty"`
						IsAnimated  bool   `json:"isAnimated,omitempty"`
						//ID          string `json:"id,omitempty"`
						Type        string `json:"type"`
						Content     string `json:"content,omitempty"`
						Modificator string `json:"modificator,omitempty"`
					} `json:"data"`
					CreatedAt int64 `json:"createdAt"`
					IsPrivate bool  `json:"isPrivate"`
					Flags     struct {
						IsParentDeleted bool `json:"isParentDeleted"`
						IsFirstMessage  bool `json:"isFirstMessage"`
					} `json:"flags"`
					//ThreadID string `json:"threadId"`
					Author User `json:"author"`
					ID     int  `json:"id"`
				} `json:"data"`
				Type string `json:"type"`
			} `json:"data"`
			Offset int `json:"offset"`
		} `json:"pub"`
	} `json:"push"`
}

func (msg *ChatMsg) GetChannelId() string {
	return strings.Split(msg.Push.Channel, ":")[1]
}

func (msg *ChatMsg) GetDisplayName() string {
	return msg.Push.Pub.Data.Data.User.DisplayName
}

func (msg *ChatMsg) GetUser() *User {
	return &msg.Push.Pub.Data.Data.User
}

//-----------

type Config struct {
	Id               string   `bson:"_id"`
	ChannelsAutoJoin []string `bson:"channelsAutoJoin"`
}

type MsgTextContent struct {
	Modificator string `json:"modificator"`
	Type        string `json:"type"`
	Content     string `json:"content"`
}

type MsgMentionContent struct {
	Type        string `json:"type"`
	ID          int    `json:"id"`
	Nick        string `json:"nick"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
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
}

// -----------
