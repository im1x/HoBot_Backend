package model

type Config struct {
	Id               string   `bson:"_id"`
	ChannelsAutoJoin []string `bson:"channelsAutoJoin"`
}
