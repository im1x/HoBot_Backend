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

type CommandsAliases struct {
	ID              string            `json:"_id"`
	CommandsAliases map[string]string `bson:"commandsAliases" json:"commandsAliases"`
}

type NewCommand struct {
	Command string `json:"command" validate:"required,gte=3"`
	Alias   string `json:"alias" validate:"required,gte=3"`
}
