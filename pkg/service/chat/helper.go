package chat

import (
	"HoBot_Backend/pkg/model"
	DB "HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service/vkplay"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

func isPING(data []byte) bool {
	if len(data) != 2 {
		return false
	}
	return data[0] == '{' && data[1] == '}'
}

func getAliasAndParamFromMessage(message string) (string, string) {
	if message == "" {
		return "", ""
	}
	commandAndParam := strings.Fields(strings.ReplaceAll(message, "\u00a0", " "))
	if len(commandAndParam) < 2 {
		commandAndParam = append(commandAndParam, "")
	}
	return commandAndParam[0], commandAndParam[1]
}
func getCommandAndPayloadForAlias(alias, channel string) (cmd, param string) {
	if chnl, ok := vkplay.ChannelsCommands.Channels[channel]; ok {
		cmd = chnl.Aliases[alias].Command
		if chnl.Aliases[alias].Payload != "" {
			param = chnl.Aliases[alias].Payload
		}
	}
	return
}

func hasAccess(alias string, msg *ChatMsg) bool {
	accessLevel := vkplay.ChannelsCommands.Channels[msg.GetChannelId()].Aliases[alias].AccessLevel
	switch accessLevel {
	case 1:
		return msg.GetUser().IsChatModerator || msg.GetUser().IsChannelModerator || msg.GetUser().IsOwner
	case 2:
		return msg.GetUser().IsOwner
	default:
		return true
	}
}

func fmtDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h == 0 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func isBotModeratorAndSentMsg(msg *ChatMsg, channelOwner model.User) bool {
	if !vkplay.IsBotHaveModeratorRights(channelOwner.Channel) {
		SendMessageToChannel("Для использования этой команды боту необходимы права модератора (для отправки личных сообщений)", msg.GetChannelId(), msg.GetUser())
		return false
	}
	return true
}

func prepareStringForSend(s string) string {
	res := strings.ReplaceAll(s, "\n", "\\n")
	res = strings.ReplaceAll(res, `"`, `\"`)
	return res
}

func GetUserNameById(ctx context.Context, id string) (string, error) {
	var user model.User
	err := DB.GetCollection(DB.Users).FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return user.Channel, err
}
