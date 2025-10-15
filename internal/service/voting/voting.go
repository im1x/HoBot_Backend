package voting

import (
	repoStatistics "HoBot_Backend/internal/repository/statistics"
	"HoBot_Backend/internal/socketio"
	"HoBot_Backend/internal/task"
	"HoBot_Backend/internal/utility"
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type VotingService struct {
	appCtx         context.Context
	Voting         map[string]*VotingData
	statisticsRepo repoStatistics.Repository
	socketioServer *socketio.SocketServer
}

func NewVotingService(appCtx context.Context, statisticsRepo repoStatistics.Repository, socketioServer *socketio.SocketServer) *VotingService {
	return &VotingService{appCtx: appCtx, Voting: make(map[string]*VotingData), statisticsRepo: statisticsRepo, socketioServer: socketioServer}
}

func (s *VotingService) StartVoting(userId string, votingRequest VotingRequest) {
	if s.Voting[userId] != nil && s.Voting[userId].StopFunc != nil {
		s.Voting[userId].StopFunc()
		delete(s.Voting, userId)
	}

	votingAnswers := make(map[string]int)
	votingResult := make(map[int]*VotingResult)

	if votingRequest.Type == 0 {
		for i, option := range votingRequest.Options {
			iPlus := i + 1
			key := strconv.Itoa(iPlus)
			votingAnswers[key] = iPlus
			votingResult[iPlus] = &VotingResult{
				Label: option,
				Count: 0,
			}
		}
		s.statisticsRepo.IncField(s.appCtx, userId, repoStatistics.Voting)
	} else {
		for i := 1; i <= 10; i++ {
			key := strconv.Itoa(i)
			votingAnswers[key] = i
		}
		s.statisticsRepo.IncField(s.appCtx, userId, repoStatistics.Rating)
	}

	if s.Voting[userId] != nil {
		s.StopVoting(userId)
	}

	stopFunc := task.CallAfterDuration(time.Minute*time.Duration(votingRequest.Duration), func() {
		s.StopVoting(userId)
	})

	s.Voting[userId] = &VotingData{
		Type:               votingRequest.Type,
		Title:              votingRequest.Title,
		IsVotingInProgress: true,
		IsHaveResult:       true,
		VotingAnswers:      votingAnswers,
		ResultVoting:       votingResult,
		ResultRating:       &RatingResult{Sum: 0, Count: 0},
		AlreadyVoted:       make(map[int]bool),
		StopAt:             votingRequest.StopAt,
		StopFunc:           stopFunc,
	}

	s.socketioServer.Emit(userId, socketio.VotingStart, s.Voting[userId].ToResponse())
}

func (s *VotingService) GetVotingStatus(userId string) VotingResponse {
	if s.Voting[userId] == nil {
		return VotingResponse{}
	}
	return s.Voting[userId].ToResponse()
}

func (s *VotingService) StopVoting(userId string) {
	if s.Voting[userId] == nil {
		return
	}
	s.Voting[userId].IsVotingInProgress = false

	if s.Voting[userId].StopFunc != nil {
		s.Voting[userId].StopFunc()
	}
	s.Voting[userId].StopFunc = task.CallAfterDuration(time.Minute*10, func() {
		delete(s.Voting, userId)
		s.socketioServer.Emit(userId, socketio.VotingDelete, "")
	})

	log.Infof("Voting stopped: %+v ", s.Voting[userId].ToResponse())

	s.socketioServer.Emit(userId, socketio.VotingStop, s.Voting[userId].ToResponse())

	// TEMP
	if s.Voting[userId].Type == 1 {
		go func() {
			ratings := s.Voting[userId].AllRatings
			log.Info("-----------------RATING-----------------")
			log.Infof("Arithmetic Mean: %.2f\n", utility.Mean(ratings))
			log.Infof("Median: %.2f\n", utility.Median(ratings))
			log.Infof("Trimmed Mean (10%% trim): %.2f\n", utility.TrimmedMean(ratings, 0.10))
			log.Infof("Bayesian Mean: %.2f\n", utility.BayesianMean(ratings, 5.5, 10))
			log.Infof("Mode-Filtered Mean (tolerance=2): %.2f\n", utility.FilterByMode(ratings, 2))
			log.Infof("Standard Deviation Filtered Mean (threshold=1.5): %.2f\n", utility.FilterByStandardDeviation(ratings, 1.5))
			log.Info("----------------------------------------")
		}()
	}
}
