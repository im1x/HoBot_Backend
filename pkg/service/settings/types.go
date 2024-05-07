package settings

import "HoBot_Backend/pkg/service/vkplay"

type SongRequestsSettings struct {
	MinVideoViews      int `bson:"min_video_views" json:"min_video_views" validate:"min=1,max=100000000"`
	MaxRequestsPerUser int `bson:"max_requests_per_user" json:"max_requests_per_user" validate:"min=0,max=30"`
	MaxDurationMinutes int `bson:"max_duration_minutes" json:"max_duration_minutes" validate:"min=0,max=60"`
}

type UserSettings struct {
	Id           string                       `bson:"_id"`
	Aliases      map[string]vkplay.CmdDetails `bson:"aliases" json:"aliases"`
	Volume       int                          `bson:"volume" json:"volume"`
	SongRequests SongRequestsSettings         `bson:"song_requests" json:"song_requests"`
}
