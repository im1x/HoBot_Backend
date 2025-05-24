package chat

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
	"time"
)

var (
	channelCache   = make(map[string]string)
	channelCacheMu sync.RWMutex
)

type VkplWs struct {
	wsConnect *websocket.Conn
	wsCounter int
	wsToken   string
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
		Achievement *struct {
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
					//ThreadID any `json:"threadId"`
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
	// Get wsID
	parts := strings.Split(msg.Push.Channel, ":")
	if len(parts) < 2 {
		log.Errorf("Invalid channel format: %s", msg.Push.Channel)
		return ""
	}
	wsID := parts[1]

	// Check cache
	channelCacheMu.RLock()
	cachedID, ok := channelCache[wsID]
	channelCacheMu.RUnlock()

	if ok {
		return cachedID
	}

	// Get channel id
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	channelId, err := GetUserIdByWs(ctx, wsID)
	if err != nil {
		log.Errorf("Error getting channel id for wsID %s: %v", wsID, err)
		return ""
	}

	// Update cache
	channelCacheMu.Lock()
	channelCache[wsID] = channelId
	channelCacheMu.Unlock()

	return channelId
}

func (msg *ChatMsg) GetDisplayName() string {
	return msg.Push.Pub.Data.Data.User.DisplayName
}

func (msg *ChatMsg) GetUser() *User {
	return &msg.Push.Pub.Data.Data.User
}

func (msg *ChatMsg) IsSubscriber() bool {
	for _, badge := range msg.GetUser().Badges {
		if badge.Achievement != nil && badge.Achievement.Type == "subscription" {
			return true
		}
	}
	return false
}

type MsgTextContent struct {
	Type        string `json:"type"`
	Content     string `json:"content"`
	Modificator string `json:"modificator"`
}

type MsgLinkContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Url     string `json:"url"`
}

type MsgMentionContent struct {
	Type        string `json:"type"`
	ID          int    `json:"id"`
	Nick        string `json:"nick"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}
