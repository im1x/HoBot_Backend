package model

import (
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Id          string `bson:"_id,omitempty" json:"id,omitempty"`
	Login       string `bson:"login,omitempty" json:"login,omitempty" validate:"required"`
	Password    string `bson:"password,omitempty" json:"password,omitempty" validate:"required,gte=8"`
	Channel     string `bson:"channel,omitempty" json:"channel,omitempty"`
	IsConfirmed bool   `bson:"confirmed,omitempty" json:"confirmed,omitempty"`
}

func (u User) ToUserDto() UserDto {
	return UserDto{
		Id:          u.Id,
		Login:       u.Login,
		IsConfirmed: u.IsConfirmed,
	}
}

type UserDto struct {
	Id          string `json:"id,omitempty"`
	Login       string `json:"login,omitempty"`
	IsConfirmed bool   `json:"confirmed"`
	jwt.RegisteredClaims
}

type UserData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         UserDto
}
