package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Login       string             `bson:"login,omitempty" json:"login,omitempty" validate:"required"`
	Password    string             `bson:"password,omitempty" json:"password,omitempty" validate:"required,gte=8"`
	Channel     string             `bson:"channel,omitempty" json:"channel,omitempty"`
	IsConfirmed bool               `bson:"confirmed,omitempty" json:"confirmed,omitempty"`
}

type UserDto struct {
	Id          primitive.ObjectID `json:"id,omitempty"`
	Login       string             `json:"login,omitempty"`
	IsConfirmed bool               `json:"confirmed"`
}

type UserData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         UserDto
}
