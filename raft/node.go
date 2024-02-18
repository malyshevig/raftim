package raft

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"raft/nw"
	"raft/raftApi"
	"time"
)

const (
	FollowerLeaderIdleBaseTimeout = 1000
	CandidateElectionTimeout      = FollowerLeaderIdleBaseTimeout
	LeaderPingInterval            = FollowerLeaderIdleBaseTimeout / 4
)

type Node struct {
	Id           int
	IncomingChan chan raftApi.MsgEvent
	OutgoingChan chan raftApi.MsgEvent
	ControlChan  chan raftApi.SystemEvent
}

type ClusterConfig struct {
	Nodes []int
}

const (
	Leader    = "leader"
	Candidate = "candidate"
	Follower  = "follower"
	Active    = "active"
)

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
	ElectionTimeoutMS int
	FollowerTimeoutMS int
}

func (t Timeout) String() string {
	return fmt.Sprintf("ElectionTimeoutMS = %d FollowerTimeoutMS = %d", t.ElectionTimeoutMS, t.FollowerTimeoutMS)
}

type RaftNode struct {
	Node
	config ClusterConfig

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
	CmdLog        []raftApi.Entry
	commitInfo    *CommitInfo
	CommitedIndex int

	Timeout

	pingNum int
	ae_id   int

	logger *zap.SugaredLogger
}

func NewNode(id int, timeouts Timeout, config ClusterConfig,
	incomingChan chan raftApi.MsgEvent,
	outgoingChan chan raftApi.MsgEvent, logger *zap.Logger) *RaftNode {

	fname := getLogName(id)
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	controlChan := make(chan raftApi.SystemEvent, nw.CHANNELSIZE)

	rn := &RaftNode{
		Node:          Node{Id: id, IncomingChan: incomingChan, OutgoingChan: outgoingChan, ControlChan: controlChan},
		config:        config,
		CurrentTerm:   1,
		State:         Follower,
		Status:        Active,
		VotedFor:      id,
		CommitedIndex: -1,
		CmdLog:        make([]raftApi.Entry, 0),
		followers:     make(map[int]*FollowerInfo),
		Timeout:       timeouts,
		logger:        logger.Sugar(),
	}
	rn.logger.Infof("Created node id: %d timeouts: %v", rn.Id, rn.Timeout)
	return rn
}

func (rn *RaftNode) init() {
	rn.CurrentTerm = 1
	rn.followerLeaderIdleTs = time.Now()
	rn.leaderPingTs = time.Now()
	rn.candidateElectionTs = time.Now()

	rn.switchToFollower(0)
}

func (rn *RaftNode) Send(msg *raftApi.MsgEvent) error {

	ch := rn.Node.OutgoingChan
	ch <- *msg
	return nil
}

func (rn *RaftNode) Run() {
	defer rn.logger.Info("node  run exit ", zap.Int("id", rn.Id))
	rn.init()
	for {

		for rn.Status != Active {
			time.Sleep(time.Duration(500) * time.Millisecond)
		}

		select {
		case controlEvent := <-rn.ControlChan:
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

		case msgEvent := <-rn.IncomingChan:
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
	rs := fmt.Sprintf("node %d vote for %d\n", rn.Id, newLeaderId)
	rn.print(rs)

	rn.VotedFor = newLeaderId
	rn.CurrentTerm = newTerm
	//ev := MsgEvent{srcid: rn.Id, dstid: newLeaderId, Body: VoteResponse{Success: true, term: newTerm}}
	ev := nw.Msg(rn.Id, newLeaderId, *raftApi.NewVoteResponse(newTerm, len(rn.CmdLog)-1, true))
	rn.Send(ev)
}

func (rn *RaftNode) SwitchToState(newState string) error {
	switch newState {
	case Leader:
		rn.switchToCandidate()
		break
	case Follower:
		rn.switchToFollower(0)
		break
	case Candidate:
		rn.switchToCandidate()
		break
	default:
		return fmt.Errorf("Incorrect state =%s\n", newState)
	}

	return nil
}

func (rn *RaftNode) switchToFollower(leaderId int) {
	rn.State = Follower
	rn.leader = LeaderInfo{id: leaderId, leaderLastTS: time.Now()}
}

func (rn *RaftNode) getFollower(srcid int) *FollowerInfo {
	return rn.followers[srcid]
}
