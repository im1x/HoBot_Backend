package songRequest

import "sync"

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
