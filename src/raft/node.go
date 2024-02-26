package raft

import (
	"fmt"
	"go.uber.org/zap"
	"raft/src/net"
	"raft/src/proto"
	"raft/src/util"
	"time"
)

const (
	FollowerLeaderIdleBaseTimeout = 1000
	CandidateElectionTimeout      = FollowerLeaderIdleBaseTimeout
	LeaderPingInterval            = FollowerLeaderIdleBaseTimeout / 4
)

type Node struct {
	Id           int
	IncomingChan chan proto.MsgEvent
	OutgoingChan chan proto.MsgEvent
	ControlChan  chan proto.SystemEvent
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
	config net.ClusterConfig

	CurrentTerm          int64
	State                string
	VotedFor             int
	VoteCount            int
	candidateElectionTs  time.Time
	followerLeaderIdleTs time.Time
	leaderPingTs         time.Time

	followers           map[int]*FollowerInfo
	followersCommitInfo CommitInfo
	leader              LeaderInfo

	Status        string
	CmdLog        []proto.Entry
	commitInfo    *CommitInfo
	CommitedIndex int

	Timeout

	ae_id  int
	logger *zap.SugaredLogger
}

func (rn *RaftNode) GetIncomingChannel() chan proto.MsgEvent {
	return rn.IncomingChan
}

func (rn *RaftNode) SetIncomingChannel(c chan proto.MsgEvent) {
	rn.IncomingChan = c
}

func (rn *RaftNode) GetOutgoingChannel() chan proto.MsgEvent {
	return rn.OutgoingChan
}

func (rn *RaftNode) SetOutgoingChannel(c chan proto.MsgEvent) {
	rn.OutgoingChan = c
}

func (rn *RaftNode) GetControlChannel() chan proto.SystemEvent {
	return rn.ControlChan
}

func (rn *RaftNode) SetControlChannel(c chan proto.SystemEvent) {
	rn.ControlChan = c
}

func (rn RaftNode) String() string {
	return fmt.Sprintf("node (%d,%s,%s) log:%d commitIndex:%d leader:%d", rn.Id, rn.State,
		rn.Status, len(rn.CmdLog), rn.CommitedIndex, rn.leader.id)
}

func NewNode(id int, config net.ClusterConfig) *RaftNode {
	loggerFile := fmt.Sprintf("logs/node%d.log", id)

	logger := util.InitLogger(loggerFile)
	tm := net.RandomiseTimeout(CandidateElectionTimeout)
	timeouts := Timeout{ElectionTimeoutMS: tm, FollowerTimeoutMS: tm}

	incomingChan := make(chan proto.MsgEvent, net.CHANNELSIZE)
	outgoingChan := make(chan proto.MsgEvent, net.CHANNELSIZE)
	controlChan := make(chan proto.SystemEvent, net.CHANNELSIZE)

	rn := &RaftNode{
		Node:          Node{Id: id, IncomingChan: incomingChan, OutgoingChan: outgoingChan, ControlChan: controlChan},
		config:        config,
		CurrentTerm:   1,
		State:         Follower,
		Status:        Active,
		VotedFor:      id,
		CommitedIndex: -1,
		CmdLog:        make([]proto.Entry, 0),
		followers:     make(map[int]*FollowerInfo),
		Timeout:       timeouts,
		logger:        logger.Sugar(),
	}
	rn.initCmdLog()
	rn.logger.Infof("%s: created node ", *rn)
	return rn
}

func (rn *RaftNode) init() {
	rn.CurrentTerm = 1
	rn.followerLeaderIdleTs = time.Now()
	rn.leaderPingTs = time.Now()
	rn.candidateElectionTs = time.Now()

	rn.switchToFollower(0)
}

func (rn *RaftNode) Send(msg *proto.MsgEvent) {
	rn.logger.Infof("%s: send msg %s", *rn, msg)
	ch := rn.Node.OutgoingChan
	ch <- *msg
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
				if rn.Status == Active {
					rn.leaderProcessSystemEvent(&controlEvent)
				}
				break
			case Follower:
				if rn.Status == Active {
					rn.followerProcessSystemEvent(&controlEvent)
				}
				break
			case Candidate:
				if rn.Status == Active {
					rn.candidateProcessSystemEvent(&controlEvent)
				}
				break
			}

		case msgEvent := <-rn.IncomingChan:
			switch rn.State {
			case Leader:
				if rn.Status == Active {
					rn.leaderProcessMsgEvent(&msgEvent)
				}
				break
			case Follower:
				if rn.Status == Active {
					rn.followerProcessMsgEvent(&msgEvent)
				}
				break
			case Candidate:
				if rn.Status == Active {
					rn.candidateProcessMsgEvent(&msgEvent)
				}
				break
			}
		}

	}
}

func (rn *RaftNode) grantVote(newLeaderId int, newTerm int64) {
	rn.logger.Infof("%s grant vote to %d", *rn, newLeaderId)

	rn.VotedFor = newLeaderId
	rn.CurrentTerm = newTerm

	ev := net.Msg(rn.Id, newLeaderId, *proto.NewVoteResponse(newTerm, len(rn.CmdLog)-1, true))
	rn.Send(ev)
}

func (rn *RaftNode) SwitchToState(newState string) error {
	rn.logger.Infof("%s switch to state %s", *rn, newState)
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
	rn.logger.Infof("%s switch to follower, leader:%d", *rn, leaderId)
	rn.State = Follower
	rn.leader = LeaderInfo{id: leaderId, leaderLastTS: time.Now()}
}

func (rn *RaftNode) getFollower(srcid int) *FollowerInfo {
	return rn.followers[srcid]
}
