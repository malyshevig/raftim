package raft

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	FollowerLeaderIdleBaseTimeout = 10
	CandidateElectionTimeout      = FollowerLeaderIdleBaseTimeout / 2
	LeaderPingInterval            = FollowerLeaderIdleBaseTimeout / 2
)

func FollowerLeaderIdleTimeout() int {
	dev := 2
	return FollowerLeaderIdleBaseTimeout - dev + rand.Intn(2*dev)
}

type Node struct {
	Id            int
	IncommingChan chan Event
}

const (
	Leader    = "leader"
	Candidate = "candidate"
	Follower  = "follower"
)

type RaftNode struct {
	Node

	CurrentTerm          int64
	State                string
	VotedFor             int
	VoteCount            int
	candidateElectionTs  time.Time
	followerLeaderIdleTs time.Time
	leaderPingTs         time.Time
}

func NewNode(id int) *RaftNode {
	return &RaftNode{
		Node:        Node{Id: id, IncommingChan: make(chan Event, 10000)},
		CurrentTerm: 1,
		State:       Follower,
		VotedFor:    id,
	}
}

func (rn *RaftNode) init() {
	rn.CurrentTerm = 1
	rn.followerLeaderIdleTs = ZeroTime()
	rn.leaderPingTs = ZeroTime()
	rn.candidateElectionTs = ZeroTime()

	rn.switchToFollower()
	//ev := Event{rn.CurrentTerm, IdleTimeout{time.Now()}}
	//go Timer(int64(RandomTimeout(5, 10)), ev, rn.Node.IncommingChan)
}

func (rn *RaftNode) nextEvent() *Event {

	ev := <-rn.IncommingChan
	return &ev
}

func (rn *RaftNode) run() {
	defer rn.print("(4) run exit")
	rn.init()
	for {
		ev := rn.nextEvent()
		if ev.Term > 0 {
			rn.print(" (5) received Event term=%d \n", ev.Term)
		}
		if ev.Term == 0 || ev.Term >= rn.CurrentTerm {
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
		} else {
			rn.print("(6) current_term= %d received_term = %d skipping", rn.CurrentTerm, ev.Term)
		}

	}

}

func (rn *RaftNode) grantVote(newLeaderId int, newTerm int64) {
	rn.VotedFor = newLeaderId
	rn.CurrentTerm = newTerm
	ev := Event{Term: rn.CurrentTerm, Inner: VoteResponse{rn.Id}}
	ClusterInstance().send(rn.Id, newLeaderId, ev)
}

func (rn *RaftNode) leaderProcessEvent(ev *Event) {
	if vr, ok := ev.Inner.(VoteRequest); ok {
		if rn.CurrentTerm < ev.Term {
			rn.switchToFollower()
			rn.grantVote(vr.srcId, ev.Term)
		}
	}

	if rn.CurrentTerm < ev.Term {
		rn.CurrentTerm = ev.Term
		rn.switchToFollower()
	}

	if _, ok := ev.Inner.(TimerTick); ok { // Idle Timeout
		if IsTimeout(rn.leaderPingTs, time.Now(), LeaderPingInterval) {
			ping := Event{rn.CurrentTerm, PingRequest{srcId: rn.Id}}
			rn.leaderPingTs = time.Now()
			ClusterInstance().sendAll(rn.Node.Id, ping)
		}
	}

	return
}

func (rn *RaftNode) followerProcessEvent(ev *Event) {

	if _, ok := ev.Inner.(TimerTick); ok { // Idle Timeout
		if IsTimeout(rn.followerLeaderIdleTs, time.Now(), FollowerLeaderIdleTimeout()) {
			rn.switchToCandidate()
			return
		}
	}

	if vr, ok := ev.Inner.(VoteRequest); ok {
		if rn.CurrentTerm < ev.Term {
			rn.switchToFollower()
			rn.grantVote(vr.srcId, ev.Term)
		}

		return
	}

	if vr, ok := ev.Inner.(PingRequest); ok {
		if rn.CurrentTerm < ev.Term {
			rn.CurrentTerm = ev.Term
			rn.VotedFor = vr.srcId
			rn.switchToFollower()
		}
		ClusterInstance().send(rn.Id, vr.srcId, Event{rn.CurrentTerm, PingResponse{srcId: rn.Id}})
		rn.followerLeaderIdleTs = time.Now()
		return
	}

	return
}

func (rn *RaftNode) candidateProcessEvent(ev *Event) {
	if _, ok := ev.Inner.(TimerTick); ok { // Idle Timeout
		if IsTimeout(rn.candidateElectionTs, time.Now(), CandidateElectionTimeout) {
			rn.switchToFollower()
			return
		}
	}

	if vr, ok := ev.Inner.(VoteRequest); ok {
		if rn.CurrentTerm < ev.Term {
			rn.CurrentTerm = ev.Term
			rn.VotedFor = vr.srcId
			rn.VoteCount = 0
			event := Event{Term: rn.CurrentTerm, Inner: VoteResponse{srcId: rn.Node.Id}}
			ClusterInstance().send(rn.Id, vr.srcId, event)
		}

		return
	}

	if vr, ok := ev.Inner.(VoteResponse); ok {
		if rn.CurrentTerm == ev.Term {
			rn.CurrentTerm = ev.Term
			rn.VoteCount++
			if rn.VoteCount > ClusterInstance().NodesCount()/2 {
				rn.switchToLeader()
			}

		} else {
			rn.print("(1) Candidate received unexpected response from %d", vr.srcId)
		}

		return
	}
}

func (rn *RaftNode) switchToCandidate() {
	rn.print("(7) switch to candidate \n")
	rn.CurrentTerm++
	rn.State = Candidate
	rn.VotedFor = rn.Id
	rn.VoteCount = 1
	event := Event{Term: rn.CurrentTerm, Inner: VoteRequest{srcId: rn.Node.Id}}
	rn.candidateElectionTs = time.Now()

	if err := ClusterInstance().sendAll(rn.Node.Id, event); err != nil {
		fmt.Println("Can't send message d", err)
	}

	//rn.timerCandidate()
}

func (rn *RaftNode) switchToLeader() {
	rn.print("(9) switch to Leader \n")

	rn.State = Leader
	rn.leaderPingTs = time.Now()

	ping := Event{rn.CurrentTerm, PingRequest{srcId: rn.Id}}
	ClusterInstance().sendAll(rn.Node.Id, ping)

	//rn.timerLeader()
}

func (rn *RaftNode) switchToFollower() {
	rn.print("(8) switch to Follower \n")
	rn.followerLeaderIdleTs = time.Now()
	rn.State = Follower
}

func (rn *RaftNode) print(fs string, params ...any) {
	fmt.Printf("Node%d (state=%s) (currentTerm = %d):  ", rn.Id, rn.State, rn.CurrentTerm)
	if len(params) > 0 {
		fmt.Printf(fs, params)
	} else {
		fmt.Print(fs)
	}
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
