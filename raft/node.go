package raft

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"
)

const (
	FollowerLeaderIdleBaseTimeout = 10000
	CandidateElectionTimeout      = FollowerLeaderIdleBaseTimeout / 2
	LeaderPingInterval            = FollowerLeaderIdleBaseTimeout / 2
)

type Node struct {
	id           int
	incomingChan chan interface{}
	outgoingChan chan interface{}
}

const (
	Leader    = "leader"
	Candidate = "candidate"
	Follower  = "follower"
	Active    = "active"
)

type Entry struct {
	term     int64
	clientId int
	msgId    int64
	cmd      string
}

type FollowerInfo struct {
	id           int
	lastRequest  time.Time
	lastResponse time.Time

	nextIndex int
}

type LeaderInfo struct {
	id           int
	leaderLastTS time.Time
}

type Timeout struct {
	electionTimeoutMS int
	followerTimeoutMS int
}

type RaftNode struct {
	Node

	CurrentTerm          int64
	State                string
	VotedFor             int
	VoteCount            int
	candidateElectionTs  time.Time
	followerLeaderIdleTs time.Time
	leaderPingTs         time.Time

	followers            map[int]*FollowerInfo
	followersCommmitInfo CommitInfo
	leader               LeaderInfo

	Status        string
	CmdLog        []Entry
	commitInfo    *CommitInfo
	commitedIndex int

	Timeout

	pingNum int
	ae_id   int
}

func NewNode(id int, r rnd, incomingChan chan interface{}, outgoingChan chan interface{}) *RaftNode {
	fname := getLogName(id)
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	return &RaftNode{
		Node:          Node{id: id, incomingChan: incomingChan, outgoingChan: outgoingChan},
		CurrentTerm:   1,
		State:         Follower,
		Status:        Active,
		VotedFor:      id,
		commitedIndex: -1,
		CmdLog:        make([]Entry, 0),
		followers:     make(map[int]*FollowerInfo),
		Timeout: Timeout{electionTimeoutMS: r.RandomTimeoutMs(CandidateElectionTimeout),
			followerTimeoutMS: r.RandomTimeoutMs(FollowerLeaderIdleBaseTimeout)},
	}
}

func (rn *RaftNode) init() {
	rn.CurrentTerm = 1
	rn.followerLeaderIdleTs = ZeroTime()
	rn.leaderPingTs = ZeroTime()
	rn.candidateElectionTs = ZeroTime()

	rn.switchToFollower()
}

func (rn *RaftNode) nextEvent() interface{} {

	// block if status inactive
	for rn.Status != Active {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	ev := <-rn.incomingChan

	types := reflect.TypeOf(ev)
	if types.String() != "raft.SystemEvent" {
		rn.print(fmt.Sprintf("received %s", reflect.TypeOf(ev)))
	}

	if msg, ok := ev.(MsgEvent); ok {
		rn.print(fmt.Sprintf(" --------- recieved msg from %d", msg.srcid))
	}
	return ev
}

func (rn *RaftNode) run() {
	defer rn.print("run exit")
	rn.init()
	for {
		ev := rn.nextEvent()

		//rn.print(fmt.Sprintf("start process event %s\n", reflect.TypeOf(ev)))
		switch rn.State {
		case Leader:
			rn.leaderProcessEvent(ev)
			break
		case Follower:
			rn.followerProcessEvent(ev)
			break
		case Candidate:
			rn.candidateProcessEvent(ev)
			break

		}

	}
}

func (rn *RaftNode) grantVote(newLeaderId int, newTerm int64) {
	rs := fmt.Sprintf("node %d vote for %d\n", rn.id, newLeaderId)
	rn.print(rs)

	rn.VotedFor = newLeaderId
	rn.CurrentTerm = newTerm
	//ev := MsgEvent{srcid: rn.id, dstid: newLeaderId, body: VoteResponse{success: true, term: newTerm}}
	ev := msg(rn.id, newLeaderId, *NewVoteResponse(newTerm, len(rn.CmdLog)-1, true))
	rn.send(ev)
}

func (rn *RaftNode) switchToState(newState string) error {
	switch newState {
	case Leader:
		rn.switchToCandidate()
		break
	case Follower:
		rn.switchToFollower()
		break
	case Candidate:
		rn.switchToCandidate()
		break
	default:
		return fmt.Errorf("Incorrect state =%s\n", newState)
	}

	return nil
}

func (rn *RaftNode) AppendCommand(cmd string) {
	rn.incomingChan <- ClientEvent{clientId: "client", body: ClientCommand{cmd: cmd}}
}

func (rn *RaftNode) switchToFollower() {
	rn.print("(8) switch to Follower \n")
	rn.followerLeaderIdleTs = time.Now()
	rn.State = Follower
}

func (rn *RaftNode) getFollower(srcid int) *FollowerInfo {
	return rn.followers[srcid]
}
