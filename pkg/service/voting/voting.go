package voting

import (
	"HoBot_Backend/pkg/socketio"
	"HoBot_Backend/pkg/task"
	"strconv"
	"time"
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
	} else {
		for i := 1; i <= 10; i++ {
			key := strconv.Itoa(i)
			votingAnswers[key] = i
		}
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
		StopAt:             time.Now().Add(time.Minute * time.Duration(votingRequest.Duration)),
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

	socketio.Emit(userId, socketio.VotingStop, Voting[userId].ToResponse())
}
