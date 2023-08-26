package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Token struct {
	UserId       primitive.ObjectID `bson:"user_id" json:"user_id"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
}
