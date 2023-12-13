package main

import (
	"HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/router"
	"HoBot_Backend/pkg/service/vkplay"
	"HoBot_Backend/pkg/socketio"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	//env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if ok := testEnvs([]string{
		"PORT",
		"MONGODB_URI",
		"DB_NAME",
		"JWT_ACCESS_SECRET",
		"JWT_REFRESH_SECRET",
		"WS_PORT",
		"VKPL_LOGIN",
		"VKPL_PASSWORD"}); !ok {
		log.Fatalln("Please add required envs")
	}

	mongo.Connect()

	ctx := context.Background()
	vkplay.ConnectWS(ctx)

	//Http server
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		MaxAge:           600,
		AllowOrigins:     "http://localhost:5173",
	}))
	router.Register(app)

	// GET /api/register
	/*	app.Get("/api/*", func(c *fiber.Ctx) error {
		msg := fmt.Sprintf("✋ %s", c.Params("*"))
		return c.SendString(msg) // => ✋ register
	})*/
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

func NewContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
