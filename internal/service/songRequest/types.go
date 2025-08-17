package songRequest

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
)

type SongRequest struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChannelId string             `json:"channel_id" bson:"channel_id"`
	By        string             `json:"by" bson:"by"`
	Requested string             `json:"requested" bson:"requested"`
	YT_ID     string             `json:"yt_id" bson:"yt_id"`
	Title     string             `json:"title" bson:"title"`
	Length    int                `json:"length" bson:"length"`
	Views     int                `json:"views" bson:"views"`
	Start     int                `json:"start" bson:"start"`
	End       int                `json:"end" bson:"end"`
}

type VotesForSkipSong struct {
	Count        int
	AlreadyVoted map[int]bool
	sync.Mutex
}

func (v *VotesForSkipSong) VoteYes(userId int) {
	v.Lock()
	defer v.Unlock()

	if !v.AlreadyVoted[userId] {
		v.AlreadyVoted[userId] = true
		v.Count++
	}
}

func (v *VotesForSkipSong) VoteNo(userId int) {
	v.Lock()
	defer v.Unlock()

	if !v.AlreadyVoted[userId] {
		v.AlreadyVoted[userId] = true
		v.Count--
	}
}

func (v *VotesForSkipSong) HasVoted(userId int) bool {
	v.Lock()
	defer v.Unlock()

	return v.AlreadyVoted[userId]
}

func (v *VotesForSkipSong) GetCount() int {
	v.Lock()
	defer v.Unlock()
	return v.Count
}
