package handler

import (
	"HoBot_Backend/pkg/service/chat"
	"HoBot_Backend/pkg/service/songRequest"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"net/url"
	"sync"
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

	var (
		wg                      sync.WaitGroup
		playlist, history       []songRequest.SongRequest
		playlistErr, historyErr error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		playlist, playlistErr = songRequest.GetPlaylist(c.Context(), userId)
	}()

	go func() {
		defer wg.Done()
		history, historyErr = songRequest.GetPlaylistHistory(c.Context(), userId)
	}()

	wg.Wait()

	if playlistErr != nil || historyErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get playlist"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"playlist": playlist, "history": history})
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

	if len(c.BodyRaw()) > 2 {
		type AutoSkip struct {
			AutoSkip bool `json:"autoSkip"`
		}
		var autoSkip AutoSkip
		err := c.BodyParser(&autoSkip)
		if err == nil && autoSkip.AutoSkip {
			song, err := songRequest.GetCurrentSong(userId)
			if err != nil {
				log.Error("Error while getting current song:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
			}
			chat.SendMessageToChannel(fmt.Sprintf("Песня \"%s\" от %s не воспроизводится, пропущена", song.Title, song.By), userId, nil)
		}
	}

	err := songRequest.SkipSong(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func ClearPlaylist(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	err := songRequest.RemoveAllSongs(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func RemoveSong(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	songId := c.Params("id")
	err := songRequest.RemoveSong(userId, songId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}
