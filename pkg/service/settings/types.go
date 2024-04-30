package settings

import "HoBot_Backend/pkg/service/vkplay"

type UserSettingsVolume struct {
	Volume int `json:"volume"`
}

// User Settings
type SongRequestSettings struct {
	MinVideoView      int `bson:"min_video_view" json:"min_video_view"`
	MaxRequestPerUser int `bson:"max_request_per_user" json:"max_request_per_user"`
}

type UserSettings struct {
	Aliases     map[string]vkplay.CmdDetails `bson:"aliases" json:"aliases"`
	Volume      int                          `bson:"volume" json:"volume"`
	SongRequest SongRequestSettings          `bson:"song_request" json:"song_request"`
}
