package chat

import (
	"encoding/json"
	"fmt"
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

func getAliasAndRestFromMessage(message string) (string, string) {
	if message == "" {
		return "", ""
	}
	commandAndRest := strings.SplitN(strings.ReplaceAll(message, "\u00a0", " "), " ", 2)
	if len(commandAndRest) < 2 {
		commandAndRest = append(commandAndRest, "")
	}
	return commandAndRest[0], commandAndRest[1]

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

func prepareStringForSend(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}

func makeContentJSON(seg string) string {
	arr := []interface{}{seg + " ", "unstyled", []interface{}{}}
	b, _ := json.Marshal(arr)
	return string(b)
}
