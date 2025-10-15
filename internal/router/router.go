package router

import (
	"HoBot_Backend/internal/handler"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, commonHandler *handler.CommonHandler, settingHandler *handler.SettingHandler, songRequestHandler *handler.SongRequestHandler, userHandler *handler.UserHandler, votingHandler *handler.VotingHandler) {
	api := app.Group("/api")

	api.Post("/logout", userHandler.Logout)
	api.Get("/refresh", userHandler.Refresh)
	api.Get("/vkpl", userHandler.VkplAuth)
	api.Get("/fCG7qDwksSthNCCczcpTXeDD/:id", commonHandler.TerminateApp)

	songRequest := api.Group("/songrequest")
	songRequest.Get("/playlist/:streamer", songRequestHandler.PlaylistByStreamer)

	app.Use(jwtware.New(jwtware.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT"})
		},
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_ACCESS_SECRET"))},
	}))

	api.Get("/user", userHandler.GetCurrentUser)
	api.Delete("/user/", userHandler.WipeUser)

	api.Post("/feedback", commonHandler.Feedback)

	settings := api.Group("/settings")
	settings.Get("/commands", settingHandler.GetCommands)
	settings.Get("/commandsdropdown", settingHandler.GetCommandsDropdown)
	settings.Post("/commands", settingHandler.AddCommandAndAlias)
	settings.Put("commands/:alias", settingHandler.EditCommand)
	settings.Delete("/commands/:alias", settingHandler.DeleteCommand)
	settings.Post("/volume/:volume", settingHandler.SaveVolume)
	settings.Get("/volume", settingHandler.GetVolume)
	settings.Post("songrequests", settingHandler.ChangeSongRequestsSettings)
	settings.Get("/songrequests", settingHandler.GetSongRequestsSettings)

	songRequest.Get("/playlist", songRequestHandler.Playlist)
	songRequest.Post("/skip", songRequestHandler.SkipSong)
	songRequest.Delete("/playlist", songRequestHandler.ClearPlaylist)
	songRequest.Delete("/playlist/:id", songRequestHandler.RemoveSong)

	voting := api.Group("/voting")
	voting.Get("/", votingHandler.GetVotingState)
	voting.Post("/start", votingHandler.StartVoting)
	voting.Post("/stop", votingHandler.StopVoting)

}
