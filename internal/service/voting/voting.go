package voting

import (
	"HoBot_Backend/internal/socketio"
	"HoBot_Backend/internal/statistics"
	"HoBot_Backend/internal/task"
	"HoBot_Backend/internal/utility"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

var Voting = make(map[string]*VotingData)

func StartVoting(userId string, votingRequest VotingRequest) {
	if Voting[userId] != nil && Voting[userId].StopFunc != nil {
		Voting[userId].StopFunc()
		delete(Voting, userId)
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
		statistics.IncField(userId, statistics.Voting)
	} else {
		for i := 1; i <= 10; i++ {
			key := strconv.Itoa(i)
			votingAnswers[key] = i
		}
		statistics.IncField(userId, statistics.Rating)
	}

	if Voting[userId] != nil {
		StopVoting(userId)
	}

	stopFunc := task.CallAfterDuration(time.Minute*time.Duration(votingRequest.Duration), func() {
		StopVoting(userId)
	})

	Voting[userId] = &VotingData{
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

	socketio.Emit(userId, socketio.VotingStart, Voting[userId].ToResponse())
}

func GetVotingStatus(userId string) VotingResponse {
	if Voting[userId] == nil {
		return VotingResponse{}
	}
	return Voting[userId].ToResponse()
}

func StopVoting(userId string) {
	if Voting[userId] == nil {
		return
	}
	Voting[userId].IsVotingInProgress = false

	if Voting[userId].StopFunc != nil {
		Voting[userId].StopFunc()
	}
	Voting[userId].StopFunc = task.CallAfterDuration(time.Minute*10, func() {
		delete(Voting, userId)
		socketio.Emit(userId, socketio.VotingDelete, "")
	})

	log.Infof("Voting stopped: %+v ", Voting[userId].ToResponse())

	socketio.Emit(userId, socketio.VotingStop, Voting[userId].ToResponse())

	// TEMP
	if Voting[userId].Type == 1 {
		go func() {
			ratings := Voting[userId].AllRatings
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
