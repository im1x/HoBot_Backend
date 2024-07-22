package model

type CommandList struct {
	ID       string `json:"_id"`
	Commands []struct {
		Group string `json:"group"`
		Items []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"items"`
	} `json:"commands"`
}

type CommandsDescription struct {
	ID                  string            `json:"_id"`
	CommandsDescription map[string]string `bson:"commandsDescription" json:"commandsDescription"`
}

type CommonCommand struct {
	Command     string `json:"command" validate:"oneof=Greating_To_User SR_PlayPause SR_SetVolume SR_SkipSong SR_SongRequest SR_CurrentSong SR_MySong Print_Text Available_Commands SR_DeleteSong SR_UsersSkipSongYes SR_UsersSkipSongNo"`
	Alias       string `json:"alias" validate:"required,gte=3"`
	AccessLevel int    `json:"access_level"`
	Description string `json:"description"`
	Payload     string `json:"payload"`
}
