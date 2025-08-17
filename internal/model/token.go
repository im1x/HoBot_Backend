package model

type Token struct {
	UserId       string `bson:"user_id" json:"user_id"`
	RefreshToken string `bson:"refresh_token" json:"refresh_token"`
}
