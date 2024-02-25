package proto

import "fmt"

type VoteRequest struct {
	Term           int64
	CommittedIndex int
}

func (v VoteRequest) String() string {
	return fmt.Sprintf("vote_request(Term:%d, CommittedIndex:%d)", v.Term, v.CommittedIndex)
}

type VoteResponse struct {
	Term         int64
	LastLogIndex int
	Success      bool
}

func NewVoteResponse(term int64, lastLogIndex int, success bool) *VoteResponse {
	return &VoteResponse{Term: term, LastLogIndex: lastLogIndex, Success: success}
}

func (v VoteResponse) String() string {
	return fmt.Sprintf("vote_response(Term:%d, Success:%v)", v.Term, v.Success)
}
