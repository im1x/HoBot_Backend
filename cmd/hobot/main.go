package main

import (
	"HoBot_Backend/internal/mongo"
	"HoBot_Backend/internal/router"
	"HoBot_Backend/internal/schedule"
	"HoBot_Backend/internal/service/chat"
	"HoBot_Backend/internal/service/settings"
	"HoBot_Backend/internal/service/vkplay"
	"HoBot_Backend/internal/socketio"
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	//env
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found")
	}
	if ok := testEnvs([]string{
		"PORT",
		"IPV6_ONLY",
		"MONGODB_URI",
		"DB_NAME",
		"JWT_ACCESS_SECRET",
		"JWT_REFRESH_SECRET",
		"WS_PORT",
		"VKPL_LOGIN",
		"VKPL_PASSWORD",
		"VKPL_APP_CREDEANTIALS",
		"BOT_VKPL_ID"}); !ok {
		log.Fatalln("Please add required envs")
	}

	mongo.Connect()

	ctx := context.Background()
	vkplay.Start(ctx)
	settings.LoadSettings()
	chat.Start()
	schedule.Start()

	//Http server
	var app *fiber.App
	if os.Getenv("IPV6_ONLY") == "true" {
		app = fiber.New(fiber.Config{
			Network: "tcp6",
		})
	} else {
		app = fiber.New()
	}

	router.Register(app)

	go socketio.Start()
	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}

func testEnvs(enums []string) bool {
	successful := true
	for _, enum := range enums {
		if _, ok := os.LookupEnv(enum); !ok {
			successful = false
			log.Printf("Env \"%s\" not found\n", enum)
		}
	}
	return successful
}
