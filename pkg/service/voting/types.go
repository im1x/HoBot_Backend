package voting

import (
	"HoBot_Backend/pkg/socketio"
	"fmt"
	"strconv"
	"time"
)

type VotingResult struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type RatingResult struct {
	Count int
	Sum   int
}

type VotingData struct {
	Type               byte
	IsVotingInProgress bool
	Title              string
	VotingAnswers      map[string]int
	AlreadyVoted       map[int]bool
	ResultVoting       map[int]*VotingResult
	ResultRating       *RatingResult
	StopAt             time.Time
	StopFunc           func()
}

func (v *VotingData) AddVote(userId int, option int) {
	//if !v.HasVoted(userID) {
	v.AlreadyVoted[userId] = true
	if v.Type == 0 {
		v.ResultVoting[option].Count += 1
	} else {
		v.ResultRating.Sum += option
		v.ResultRating.Count += 1
	}
	fmt.Println("User ", userId, " voted for option ", option)
	fmt.Println(v.ResultVoting)
	//return
	//}
	//fmt.Println("User ", userID, " already voted")

	socketio.Emit(strconv.Itoa(userId), socketio.VotingVote, fmt.Sprintf("User ", userId, " voted for option ", option))
	fmt.Println(v.ResultVoting)

}

func (v *VotingData) HasVoted(userID int) bool {
	return v.AlreadyVoted[userID]
}

type VotingRequest struct {
	Type     byte
	Title    string   `json:"title"`
	Duration int      `json:"duration"`
	Options  []string `json:"options"`
}

type VotingResponse struct {
	Type               byte           `json:"type"`
	IsVotingInProgress bool           `json:"isVotingInProgress"`
	Title              string         `json:"title"`
	ResultVoting       []VotingResult `json:"resultVoting"`
	ResultRating       int            `json:"resultRating"`
	StopAt             string         `json:"stopAt"`
}

func (v *VotingData) ToResponse() VotingResponse {
	votingResult := make([]VotingResult, 0, len(v.ResultVoting))

	//for _, result := range v.ResultVoting {
	for i := 1; i <= len(v.ResultVoting); i++ {
		votingResult = append(votingResult, *v.ResultVoting[i])
	}

	rating := 0
	if v.ResultRating.Count != 0 {
		rating = v.ResultRating.Sum / v.ResultRating.Count
	}

	return VotingResponse{
		Type:               v.Type,
		IsVotingInProgress: v.IsVotingInProgress,
		Title:              v.Title,
		ResultVoting:       votingResult,
		ResultRating:       rating,
		StopAt:             v.StopAt.Format(time.RFC3339),
	}
}
