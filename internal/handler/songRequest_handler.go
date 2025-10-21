package handler

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/repository/songrequests"
	repoSongRequests "HoBot_Backend/internal/repository/songrequests"
	repoSongRequestsHistory "HoBot_Backend/internal/repository/songrequestshistory"
	repoStatistics "HoBot_Backend/internal/repository/statistics"
	repoUser "HoBot_Backend/internal/repository/user"
	"HoBot_Backend/internal/service/chat"
	"HoBot_Backend/internal/service/songRequest"
	"HoBot_Backend/internal/service/vkplay"
	"fmt"
	"net/url"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type SongRequestHandler struct {
	songrequestsRepo        repoSongRequests.Repository
	userRepo                repoUser.Repository
	songrequestsHistoryRepo repoSongRequestsHistory.Repository
	statisticsRepo          repoStatistics.Repository
	songrequestsService     *songRequest.SongRequestService
	chatService             *chat.ChatService
}

func NewSongRequestHandler(songrequestsRepo songrequests.Repository, userRepo repoUser.Repository, songrequestsHistoryRepo repoSongRequestsHistory.Repository, statisticsRepo repoStatistics.Repository, songrequestsService *songRequest.SongRequestService, chatService *chat.ChatService) *SongRequestHandler {
	return &SongRequestHandler{
		songrequestsRepo:        songrequestsRepo,
		userRepo:                userRepo,
		songrequestsHistoryRepo: songrequestsHistoryRepo,
		statisticsRepo:          statisticsRepo,
		songrequestsService:     songrequestsService,
		chatService:             chatService}
}

func (s *SongRequestHandler) PlaylistByStreamer(c *fiber.Ctx) error {
	streamer, err := url.QueryUnescape(c.Params("streamer"))
	if streamer == "" || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Streamer is empty")
	}

	userIdWs, err := vkplay.GetUserIdWsByName(streamer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	userId, err := s.userRepo.GetUserIdByWs(c.Context(), userIdWs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	var (
		wg                      sync.WaitGroup
		playlist, history       []model.SongRequest
		playlistErr, historyErr error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		playlist, playlistErr = s.songrequestsRepo.GetPlaylist(c.Context(), userId)
	}()

	go func() {
		defer wg.Done()
		history, historyErr = s.songrequestsHistoryRepo.GetPlaylistHistory(c.Context(), userId)
	}()

	wg.Wait()

	if playlistErr != nil || historyErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get playlist"})
	}

	s.statisticsRepo.IncField(c.Context(), userId, repoStatistics.SongRequestsShowPublicPlaylist)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"playlist": playlist, "history": history})
}

func (s *SongRequestHandler) Playlist(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	playlist, err := s.songrequestsRepo.GetPlaylist(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(playlist)
}

func (s *SongRequestHandler) SkipSong(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)

	if len(c.BodyRaw()) > 2 {
		type AutoSkip struct {
			AutoSkip bool `json:"autoSkip"`
		}
		var autoSkip AutoSkip
		err := c.BodyParser(&autoSkip)
		if err == nil && autoSkip.AutoSkip {
			song, err := s.songrequestsRepo.GetCurrentSong(c.Context(), userId)
			if err != nil {
				log.Error("Error while getting current song:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
			}
			s.chatService.SendMessageToChannel(fmt.Sprintf("Песня \"%s\" от %s не воспроизводится, пропущена", song.Title, song.By), userId, nil)
		}
	}

	err := s.songrequestsService.SkipSong(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	s.statisticsRepo.IncField(c.Context(), userId, repoStatistics.SongRequestsPlayed)
	return nil
}

func (s *SongRequestHandler) ClearPlaylist(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	err := s.songrequestsRepo.RemoveAllSongs(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func (s *SongRequestHandler) RemoveSong(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	songId := c.Params("id")
	err := s.songrequestsRepo.RemoveSong(c.Context(), userId, songId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}
