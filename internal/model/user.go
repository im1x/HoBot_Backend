package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Id        string `bson:"_id,omitempty" json:"id,omitempty"`
	Nick      string `bson:"nick,omitempty" json:"nick,omitempty"`
	Channel   string `bson:"channel,omitempty" json:"channel,omitempty"`
	ChannelWS string `bson:"channel_ws,omitempty"`
	AvatarURL string `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
}

func (u User) ToUserDto() UserDto {
	return UserDto{
		Id:      u.Id,
		Channel: u.Channel,
	}
}

type UserDto struct {
	Id      string `json:"id,omitempty"`
	Channel string `json:"channel,omitempty"`
	jwt.RegisteredClaims
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

type CurrentUserVkpl struct {
	Data struct {
		User struct {
			ID                 int    `json:"id"`
			Nick               string `json:"nick"`
			NickColor          int    `json:"nick_color"`
			AvatarURL          string `json:"avatar_url"`
			IsStreamer         bool   `json:"is_streamer"`
			IsVerifiedStreamer bool   `json:"is_verified_streamer"`
		} `json:"user"`
		Channel struct {
			Url string `json:"url"`
		} `json:"channel"`
	} `json:"data"`
}
