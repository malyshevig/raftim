package raft

import (
	"fmt"
	"raft/nw"
	"raft/raftApi"
	"reflect"
	"time"
)

func (rn *RaftNode) candidateProcessEvent(ev any) {

	if se, ok := ev.(raftApi.SystemEvent); ok {
		rn.candidateProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(raftApi.MsgEvent); ok {
		rn.candidateProcessMsgEvent(&msg)
		return
	}

	rs := fmt.Sprintf("unexpected Event type %s\n", reflect.TypeOf(ev))
	rn.print(rs)
	return
}

func (rn *RaftNode) switchToCandidate() {
	rn.print("switch to candidate \n")
	rn.CurrentTerm++
	rn.State = Candidate
	rn.VotedFor = rn.Id
	rn.VoteCount = 1

	rn.followers = rn.makeFollowers()

	rn.sendVoteRequest(rn.followers)
	rn.candidateElectionTs = time.Now()

}

func (rn *RaftNode) sendVoteRequest(followers map[int]*FollowerInfo) {

	for _, f := range followers {
		m := nw.Msg(rn.Id, f.id, raftApi.VoteRequest{Term: rn.CurrentTerm, CommittedIndex: rn.CommitedIndex})
		rn.Send(m)
	}

}

func (rn *RaftNode) candidateProcessSystemEvent(se *raftApi.SystemEvent) {
	//rn.print(fmt.Sprintf("candidate event %d %d\n", rn.candidateElectionTs, rn.ElectionTimeoutMS))
	if _, ok := se.Body.(raftApi.TimerTick); ok { // Idle Timeout
		if nw.IsTimeout(rn.candidateElectionTs, time.Now(), rn.ElectionTimeoutMS) {
			rn.print("reinit election")
			rn.switchToCandidate()

		}
	}
}

func (rn *RaftNode) candidateProcessMsgEvent(message *raftApi.MsgEvent) {
	rn.print(fmt.Sprintf("received msg = %v", message.Body))

	if vr, ok := message.Body.(raftApi.VoteResponse); ok {
		vf := rn.getFollower(message.Srcid)
		rn.print(fmt.Sprintf("vote response src=%d term=%d \n", message.Srcid, vr.Term))

		if vf != nil {
			vf.lastResponse = time.Now()
			vf.nextIndex = vr.LastLogIndex + 1
		}

		rn.VoteCount++
		fn := len(rn.followers)

		if rn.VoteCount > (fn+1)/2 {
			rn.switchToLeader()
		}
		return
	}

	if vr, ok := message.Body.(raftApi.VoteRequest); ok {
		rn.print(fmt.Sprintf("vote request src=%d term=%d \n", message.Srcid, vr.Term))
		if rn.CurrentTerm < vr.Term {
			rn.switchToFollower(message.Srcid)
			rn.CurrentTerm = vr.Term

			if rn.CommitedIndex <= vr.CommittedIndex {
				rn.grantVote(message.Srcid, vr.Term)
				rn.VotedFor = 0

			}
		}
	}

	if ar, ok := message.Body.(raftApi.AppendEntries); ok {
		if rn.CurrentTerm <= ar.Term {
			rn.CurrentTerm = ar.Term

			rn.switchToFollower(message.Srcid)
			rn.followerProcessMsgEvent(message)
		}
		return
	}

	if cmd, ok := message.Body.(raftApi.ClientCommand); ok {
		rn.Send(nw.Msg(rn.Id, message.Srcid, raftApi.ClientCommandResponse{CmdId: cmd.Id, Success: false, Leaderid: rn.leader.id}))
		return
	}
	if _, ok := message.Body.(raftApi.LeaderDiscoveryRequest); ok {
		//rn.Send(Msg(rn.Id, message.srcid, LeaderDiscoveryResponse{leaderId: 0}))
		return
	}

}
