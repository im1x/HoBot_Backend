package songRequest

import (
	"HoBot_Backend/internal/service/settings"
	"context"
	"time"

	repoSongRequests "HoBot_Backend/internal/repository/songrequests"
	repoSongRequestsHistory "HoBot_Backend/internal/repository/songrequestshistory"

	"github.com/gofiber/fiber/v2/log"
)

type SongRequestService struct {
	ctxApp                  context.Context
	VotesForSkip            map[string]*VotesForSkipSong
	songRequestsRepo        repoSongRequests.Repository
	songRequestsHistoryRepo repoSongRequestsHistory.Repository
	settingsService         settings.SettingsService
}

func NewSongRequestService(ctxApp context.Context, songRequestsRepo repoSongRequests.Repository, songRequestsHistoryRepo repoSongRequestsHistory.Repository, settingsService *settings.SettingsService) *SongRequestService {
	return &SongRequestService{
		ctxApp:                  ctxApp,
		VotesForSkip:            make(map[string]*VotesForSkipSong),
		songRequestsRepo:        songRequestsRepo,
		songRequestsHistoryRepo: songRequestsHistoryRepo,
		settingsService:         *settingsService,
	}
}

func (s *SongRequestService) SkipSong(channelId string) error {
	ctx, cancel := context.WithTimeout(s.ctxApp, 3*time.Second)
	defer cancel()

	song, err := s.songRequestsRepo.SkipSong(ctx, channelId)
	if err != nil {
		log.Error("Error while deleting song request:", err)
		return err
	}

	err = s.songRequestsHistoryRepo.SaveSongRequestToHistory(song)
	if err != nil {
		return err
	}

	if s.settingsService.UsersSettings[channelId].SongRequests.IsUsersSkipAllowed {
		s.VotesForSkip[channelId] = nil
	}

	return nil
}

func (s *SongRequestService) InitUsersSkipIfNeeded(userId string) {
	if s.VotesForSkip[userId] == nil {
		s.VotesForSkip[userId] = &VotesForSkipSong{
			AlreadyVoted: make(map[int]bool),
		}
	}
}

func (s *SongRequestService) VotesForSkipYes(channelId string, userId int) bool {
	s.InitUsersSkipIfNeeded(channelId)
	s.VotesForSkip[channelId].VoteYes(userId)

	if s.VotesForSkip[channelId].GetCount() >= s.settingsService.UsersSettings[channelId].SongRequests.UsersSkipValue {
		err := s.SkipSong(channelId)
		return err == nil
	}

	return false
}

func (s *SongRequestService) VotesForSkipNo(channelId string, userId int) {
	s.InitUsersSkipIfNeeded(channelId)
	s.VotesForSkip[channelId].VoteNo(userId)
}
