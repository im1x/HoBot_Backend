package handler

import (
	"HoBot_Backend/pkg/service/songRequest"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func PlaylistByStreamer(c *fiber.Ctx) error {
	streamer, err := url.QueryUnescape(c.Params("streamer"))
	if streamer == "" || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Streamer is empty")
	}

	userId, err := songRequest.GetUserIdByName(streamer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	playlist, err := songRequest.GetPlaylist(c.Context(), userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(playlist)
}

func Playlist(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	playlist, err := songRequest.GetPlaylist(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(playlist)
}

func SkipSong(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	err := songRequest.SkipSong(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}
	return nil
}
