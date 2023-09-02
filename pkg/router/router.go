package router

import (
	userHandler "HoBot_Backend/pkg/handler"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"os"
)

func Register(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/register", userHandler.Register)
	api.Post("/login", userHandler.Login)
	api.Post("/logout", userHandler.Logout)
	api.Get("/refresh", userHandler.Refresh)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_ACCESS_SECRET"))},
	}))

	api.Get("/users", userHandler.Users)
}
