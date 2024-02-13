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

func (v ClientCommand) String() string {
	return fmt.Sprintf("cmd(id: %d, cmd:%s )", v.id, v.cmd)
}

func NewClientCommand(cmd string) *ClientCommand {
	return &ClientCommand{cmd: cmd}
}

type ClientCommandResponse struct {
	cmdId   int64
	success bool

	leaderid int
}

func (v ClientCommandResponse) String() string {
	return fmt.Sprintf("cmd_resp(id: %d, success:%v leaderid:%d )", v.cmdId, v.success, v.leaderid)
}

func NewClientCommendResponse(success bool) *ClientCommandResponse {
	return &ClientCommandResponse{success: success}
}

type LeaderDiscoveryRequest struct {
}

type LeaderDiscoveryResponse struct {
	leaderId int
}
