package songRequest

type SongRequest struct {
	ChannelId string `json:"channel_id" bson:"channel_id"`
	By        string `json:"by" bson:"by"`
	Requested string `json:"requested" bson:"requested"`
	YT_ID     string `json:"yt_id" bson:"yt_id"`
	Title     string `json:"title" bson:"title"`
	Length    int    `json:"length" bson:"length"`
	Views     int    `json:"views" bson:"views"`
	Start     int    `json:"start" bson:"start"`
	End       int    `json:"end" bson:"end"`
}
