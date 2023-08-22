package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Login     string             `bson:"login,omitempty" json:"login,omitempty"`
	Password  string             `bson:"password,omitempty" json:"password,omitempty"`
	Confirmed bool               `bson:"confirmed,omitempty" json:"confirmed,omitempty"`
}
