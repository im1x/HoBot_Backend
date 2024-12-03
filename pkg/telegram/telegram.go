package telegram

import (
	"github.com/gofiber/fiber/v2/log"
	"net/http"
	"os"
)

func SendMessage(msg string) {
	go func() {
		_, err := http.Post("https://api.telegram.org/bot"+
			os.Getenv("TELEGRAM_BOT_TOKEN")+
			"/sendMessage?chat_id=-4743294690&text="+
			msg, "application/json", nil)
		if err != nil {
			log.Error("Error while sending message to telegram:", err)
			return
		}
	}()
}
