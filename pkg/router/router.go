package router

import (
	"HoBot_Backend/pkg/handler"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"os"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/logout", handler.Logout)
	api.Get("/refresh", handler.Refresh)
	api.Get("/vkpl", handler.VkplAuth)
	api.Get("/fCG7qDwksSthNCCczcpTXeDD/:id", handler.TerminateApp)

	songRequest := api.Group("/songrequest")
	songRequest.Get("/playlist/:streamer", handler.PlaylistByStreamer)

	app.Use(jwtware.New(jwtware.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT"})
		},
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_ACCESS_SECRET"))},
	}))

	api.Get("/user", handler.GetCurrentUser)
	api.Delete("/user/", handler.WipeUser)

	api.Post("/feedback", handler.Feedback)

	settings := api.Group("/settings")
	settings.Get("/commands", handler.GetCommands)
	settings.Get("/commandsdropdown", handler.GetCommandsDropdown)
	settings.Post("/commands", handler.AddCommandAndAlias)
	settings.Put("commands/:alias", handler.EditCommand)
	settings.Delete("/commands/:alias", handler.DeleteCommand)

	songRequest.Get("/playlist", handler.Playlist)
	songRequest.Post("/skip", handler.SkipSong)
	songRequest.Delete("/playlist", handler.ClearPlaylist)
	songRequest.Delete("/playlist/:id", handler.RemoveSong)

}
