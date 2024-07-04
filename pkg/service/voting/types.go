package voting

import (
	"HoBot_Backend/pkg/socketio"
	"github.com/gofiber/fiber/v2/log"
	"sync"
)

type VotingResult struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type RatingResult struct {
	Count int `json:"count"`
	Sum   int `json:"sum"`
}

type Vote struct {
	Name string `json:"name"`
	Vote int    `json:"vote"`
}

type VotingRequest struct {
	Type     byte     `json:"type" validate:"oneof=0 1"`
	Title    string   `json:"title"`
	Duration int      `json:"duration" validate:"min=1,max=60"`
	StopAt   string   `json:"stopAt" validate:"required"`
	Options  []string `json:"options"`
}

type VotingResponse struct {
	Type               byte           `json:"type"`
	IsVotingInProgress bool           `json:"isVotingInProgress"`
	IsHaveResult       bool           `json:"isHaveResult"`
	Title              string         `json:"title"`
	ResultVoting       []VotingResult `json:"resultVoting"`
	ResultRating       RatingResult   `json:"resultRating"`
	StopAt             string         `json:"stopAt"`
}

type VotingData struct {
	sync.Mutex
	Type               byte
	IsVotingInProgress bool
	IsHaveResult       bool
	Title              string
	VotingAnswers      map[string]int
	AlreadyVoted       map[int]bool
	ResultVoting       map[int]*VotingResult
	ResultRating       *RatingResult
	StopAt             string
	StopFunc           func()
}

func (v *VotingData) AddVote(channelId string, userId int, userName string, option int) {
	if !v.HasVoted(userId) {
		v.Lock()
		v.AlreadyVoted[userId] = true
		if v.Type == 0 {
			v.ResultVoting[option].Count += 1
		} else {
			v.ResultRating.Sum += option
			v.ResultRating.Count += 1
			log.Infof("%s voted for %d ( %d / %d) ", userName, option, v.ResultRating.Count, v.ResultRating.Sum)
		}
		v.Unlock()
		socketio.Emit(channelId, socketio.VotingVote, &Vote{Name: userName, Vote: option})
	}
}

func (v *VotingData) HasVoted(userID int) bool {
	return v.AlreadyVoted[userID]
}

func (v *VotingData) ToResponse() VotingResponse {
	votingResult := make([]VotingResult, 0, len(v.ResultVoting))
	if v.Type == 0 {
		for i := 1; i <= len(v.ResultVoting); i++ {
			votingResult = append(votingResult, *v.ResultVoting[i])
		}
	}

	return VotingResponse{
		Type:               v.Type,
		IsVotingInProgress: v.IsVotingInProgress,
		IsHaveResult:       v.IsHaveResult,
		Title:              v.Title,
		ResultVoting:       votingResult,
		ResultRating:       *v.ResultRating,
		StopAt:             v.StopAt,
	}
}
