package raftApi

import (
	"fmt"
	"time"
)

type SystemEvent struct {
	Body interface{}
}

type MsgEvent struct {
	Srcid int
	Dstid int

	Ts   time.Time
	Body interface{}
}

// System events

type TimerTick struct {
}

// MsgEvent

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

type Entry struct {
	Term     int64
	ClientId int
	MsgId    int64
	Cmd      string
}

type AppendEntries struct {
	Id      int
	Term    int64
	Entries []Entry

	LastLogIndex int
	LastLogTerm  int64

	LeaderCommittedIndex int
}

func (v AppendEntries) String() string {
	return fmt.Sprintf("ae(Id:%d, Term:%d, len(Entries):%d, laslogIndex: %d, lastlogTerm: %d, LeaderCommittedIndex: %d)",
		v.Id, v.Term, len(v.Entries), v.LastLogIndex, v.LastLogTerm, v.LeaderCommittedIndex)
}

type AppendEntriesResponse struct {
	Ae_id     int
	Success   bool
	LastIndex int
}

func (v AppendEntriesResponse) String() string {
	return fmt.Sprintf("ae_resp(ae_id: %d, Success:%v, lastIndex:%d)", v.Ae_id, v.Success, v.LastIndex)
}

// ClientEvents
type ClientCommand struct {
	Id  int64
	Cmd string
}

func (v ClientCommand) String() string {
	return fmt.Sprintf("Cmd(Id: %d, Cmd:%s )", v.Id, v.Cmd)
}

func NewClientCommand(cmd string) *ClientCommand {
	return &ClientCommand{Cmd: cmd}
}

type ClientCommandResponse struct {
	CmdId   int64
	Success bool

	Leaderid int
}

func (v ClientCommandResponse) String() string {
	return fmt.Sprintf("cmd_resp(Id: %d, Success:%v Leaderid:%d )", v.CmdId, v.Success, v.Leaderid)
}

func NewClientCommendResponse(success bool) *ClientCommandResponse {
	return &ClientCommandResponse{Success: success}
}

type LeaderDiscoveryRequest struct {
}

func (v LeaderDiscoveryRequest) String() string {
	return fmt.Sprintf("leader_req()")
}

type LeaderDiscoveryResponse struct {
	LeaderId int
}

func (v LeaderDiscoveryResponse) String() string {
	return fmt.Sprintf("leader_resp(LeaderId: %d)", v.LeaderId)
}
