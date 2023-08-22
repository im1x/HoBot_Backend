package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	//env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if ok := testEnvs([]string{"PORT", "DB_URL", "DB_NAME"}); !ok {
		log.Fatalln("Please add required envs")
	}

	//Http server
	app := fiber.New()

	// GET /api/register
	app.Get("/api/*", func(c *fiber.Ctx) error {
		msg := fmt.Sprintf("✋ %s", c.Params("*"))
		return c.SendString(msg) // => ✋ register
	})
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
