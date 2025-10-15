package chat

import (
	"github.com/gorilla/websocket"
)

type VkplWs struct {
	wsConnect *websocket.Conn
	wsCounter int
	wsToken   string
}

// -----------

type MsgTextContent struct {
	Type        string `json:"type"`
	Content     string `json:"content"`
	Modificator string `json:"modificator"`
}

type MsgLinkContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Url     string `json:"url"`
}

type MsgMentionContent struct {
	Type        string `json:"type"`
	ID          int    `json:"id"`
	Nick        string `json:"nick"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

type ErrorResponse struct {
	ErrorDescription string `json:"error_description"`
	Error            string `json:"error"`
	Data             struct {
		Reasons []struct {
			RemainingTime int    `json:"remainingTime"`
			Type          string `json:"type"`
		} `json:"reasons"`
	} `json:"data"`
}
