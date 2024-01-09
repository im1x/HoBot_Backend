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
	Command     string `json:"command" validate:"required,gte=3"`
	Alias       string `json:"alias" validate:"required,gte=3"`
	AccessLevel int    `json:"access_level"`
	Description string `json:"description"`
}
