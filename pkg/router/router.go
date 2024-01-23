package router

import (
	"HoBot_Backend/pkg/handler"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"os"
)

func Register(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)
	api.Post("/logout", handler.Logout)
	api.Get("/refresh", handler.Refresh)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_ACCESS_SECRET"))},
	}))

	api.Get("/users", handler.Users)

	settings := api.Group("/settings")
	settings.Get("/commands", handler.GetCommands)
	settings.Get("/commandsdropdown", handler.GetCommandsDropdown)
	settings.Post("/commands", handler.AddCommandAndAlias)
	settings.Put("commands/:alias", handler.EditCommand)
	settings.Delete("/commands/:alias", handler.DeleteCommand)
}
