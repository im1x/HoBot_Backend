package model

type Feedback struct {
	UserId  string `json:"userId" bson:"userId"`
	Text    string `json:"text" bson:"text"`
	AddedAt string `json:"addedAt" bson:"addedAt"`
}
