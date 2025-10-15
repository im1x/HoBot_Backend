package mongodb

import "go.mongodb.org/mongo-driver/v2/mongo"

type CollectionName string

const (
	Users               CollectionName = "Users"
	Tokens                             = "Tokens"
	Vkpl                               = "Vkpl"
	Config                             = "Config"
	SettingsOptions                    = "SettingsOptions"
	UserSettings                       = "UserSettings"
	SongRequests                       = "SongRequests"
	SongRequestsHistory                = "SongRequestsHistory"
	Feedback                           = "Feedback"
	Statistics                         = "Statistics"
	PrivilegedLasqaKp                  = "PrivilegedLasqaKp"
)

func (c *Client) GetCollection(col CollectionName) *mongo.Collection {
	return c.client.Database(c.database).Collection(string(col))
}
