package model

type Token struct {
	User         User   `bson:"user,omitempty" json:"user,omitempty"`
	RefreshToken string `bson:"refresh_token,omitempty" json:"refresh_token,omitempty"`
}
