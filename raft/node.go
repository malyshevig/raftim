package raft

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	FollowerLeaderIdleBaseTimeout = 1000
	CandidateElectionTimeout      = FollowerLeaderIdleBaseTimeout / 2
	LeaderPingInterval            = FollowerLeaderIdleBaseTimeout / 2
)

type Node struct {
	id           int
	incomingChan chan MsgEvent
	outgoingChan chan MsgEvent
	controlChan  chan SystemEvent
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

func NewNode(id int, r rnd, incomingChan chan MsgEvent, outgoingChan chan MsgEvent) *RaftNode {
	fname := getLogName(id)
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	controlChan := make(chan SystemEvent, CHANNELSIZE)

	return &RaftNode{
		Node:          Node{id: id, incomingChan: incomingChan, outgoingChan: outgoingChan, controlChan: controlChan},
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
	rn.followerLeaderIdleTs = time.Now()
	rn.leaderPingTs = time.Now()
	rn.candidateElectionTs = time.Now()

	rn.switchToFollower()
}

func (rn *RaftNode) run() {
	defer rn.print("run exit")
	rn.init()
	for {

		for rn.Status != Active {
			time.Sleep(time.Duration(500) * time.Millisecond)
		}

		select {
		case controlEvent := <-rn.controlChan:
			switch rn.State {
			case Leader:
				rn.leaderProcessSystemEvent(&controlEvent)
				break
			case Follower:
				rn.followerProcessSystemEvent(&controlEvent)
				break
			case Candidate:
				rn.candidateProcessSystemEvent(&controlEvent)
				break
			}

		case msgEvent := <-rn.incomingChan:
			switch rn.State {
			case Leader:
				rn.leaderProcessMsgEvent(&msgEvent)
				break
			case Follower:
				rn.followerProcessMsgEvent(&msgEvent)
				break
			case Candidate:
				rn.candidateProcessMsgEvent(&msgEvent)
				break
			}
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

func (rn *RaftNode) switchToFollower() {
	rn.State = Follower
}

func (rn *RaftNode) getFollower(srcid int) *FollowerInfo {
	return rn.followers[srcid]
}
