package raft

import (
	"fmt"
	"time"
)

type SystemEvent struct {
	body interface{}
}

type MsgEvent struct {
	srcid int
	dstid int

	ts   time.Time
	body interface{}
}

type ClientEvent struct {
	clientId string

	body interface{}
}

// System events

type TimerTick struct {
}

// MsgEvent

type VoteRequest struct {
	term           int64
	committedIndex int
}

func (v VoteRequest) String() string {
	return fmt.Sprintf("vote_request(term:%d, committedIndex:%d)", v.term, v.committedIndex)
}

type VoteResponse struct {
	term         int64
	lastLogIndex int
	success      bool
}

func NewVoteResponse(term int64, lastLogIndex int, success bool) *VoteResponse {
	return &VoteResponse{term: term, lastLogIndex: lastLogIndex, success: success}
}

func (v VoteResponse) String() string {
	return fmt.Sprintf("vote_response(term:%d, success:%v)", v.term, v.success)
}

type AppendEntries struct {
	id      int
	term    int64
	entries []Entry

	lastLogIndex int
	lastLogTerm  int64

	leaderCommittedIndex int
}

func (v AppendEntries) String() string {
	return fmt.Sprintf("ae(id:%d, term:%d, len(entries):%d, laslogIndex: %d, lastlogTerm: %d, leaderCommittedIndex: %d)",
		v.id, v.term, len(v.entries), v.lastLogIndex, v.lastLogTerm, v.leaderCommittedIndex)
}

type AppendEntriesResponse struct {
	ae_id     int
	success   bool
	lastIndex int
}

func (v AppendEntriesResponse) String() string {
	return fmt.Sprintf("ae_resp(ae_id: %d, success:%v, lastIndex:%d)", v.ae_id, v.success, v.lastIndex)
}

// ClientEvents
type ClientCommand struct {
	id  int64
	cmd string
}

func NewClientCommand(cmd string) *ClientCommand {
	return &ClientCommand{cmd: cmd}
}

type ClientCommendResponse struct {
	cmdId   int64
	success bool

	leaderid int
}

func NewClientCommendResponse(success bool) *ClientCommendResponse {
	return &ClientCommendResponse{success: success}
}
