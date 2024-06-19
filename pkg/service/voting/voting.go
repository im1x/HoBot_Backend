package voting

import (
	"HoBot_Backend/pkg/socketio"
	"HoBot_Backend/pkg/task"
	"fmt"
	"strconv"
	"time"
)

var Voting = make(map[string]*VotingData)

func StartVoting(userId string, votingRequest VotingRequest) {
	if Voting[userId] != nil && Voting[userId].StopFunc != nil {
		fmt.Println("StopFunc on start voting")
		Voting[userId].StopFunc()
		delete(Voting, userId)
	}

	votingAnswers := make(map[string]int)
	votingResult := make(map[int]*VotingResult)
	for i, option := range votingRequest.Options {
		iPlus := i + 1
		//for i := 1; i <= len(votingRequest.Options); i++ {
		key := strconv.Itoa(iPlus)
		votingAnswers[key] = iPlus
		votingResult[iPlus] = &VotingResult{
			Label: option,
			Count: 0,
		}
	}

	if Voting[userId] != nil {
		StopVoting(userId)
	}

	stopFunc := task.CallAfterDuration(time.Minute*time.Duration(votingRequest.Duration), func() {
		fmt.Println("callback")
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

	if Voting[userId].Type == 0 {
		socketio.Emit(userId, socketio.VotingStart, Voting[userId].ToResponse())
	} else {
		socketio.Emit(userId, socketio.VotingStart, "rating")
	}

	fmt.Println("Voting for user ", userId, " started")
	fmt.Println(Voting[userId])

}

func GetVotingStatus(userId string) VotingResponse {
	if Voting[userId] == nil {
		return VotingResponse{}
	}
	return Voting[userId].ToResponse()
}

func StopVoting(userId string) {
	if Voting[userId] == nil {
		fmt.Println("-----VOTING IS NULL-----")
		return
	}
	//if Voting[userId] != nil {
	Voting[userId].IsVotingInProgress = false
	//}

	/*if data, ok := Voting[userId]; ok {
		data.IsVotingInProgress = false
		Voting[userId] = data
	}*/

	if Voting[userId].StopFunc != nil {
		fmt.Println("StopFunc")
		Voting[userId].StopFunc()
	}
	Voting[userId].StopFunc = task.CallAfterDuration(time.Minute*1, func() {
		fmt.Println("Delete voting")
		delete(Voting, userId)
		//Voting[userId] = nil
		socketio.Emit(userId, socketio.VotingDelete, "")
	})

	socketio.Emit(userId, socketio.VotingStop, Voting[userId].ToResponse())

	fmt.Println("Voting for user ", userId, " stopped")
}
