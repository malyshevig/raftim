package raft

import (
	"fmt"
	"reflect"
	"time"
)

func (rn *RaftNode) candidateProcessEvent(ev any) {

	if se, ok := ev.(SystemEvent); ok {
		rn.candidateProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(MsgEvent); ok {
		rn.candidateMsgEvent(&msg)
		return
	}

	if cm, ok := ev.(ClientEvent); ok {
		rn.candidateClientEvent(&cm)
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
	rn.VotedFor = rn.id
	rn.VoteCount = 1

	rn.followers = rn.makeFollowers()

	rn.sendVoteRequest(rn.followers)
	rn.candidateElectionTs = time.Now()

}

func (rn *RaftNode) sendVoteRequest(followers map[int]*FollowerInfo) {

	for _, f := range followers {
		m := msg(rn.id, f.id, VoteRequest{term: rn.CurrentTerm, committedIndex: rn.commitedIndex})
		rn.send(m)
	}

}

func (rn *RaftNode) candidateProcessSystemEvent(se *SystemEvent) {
	//rn.print(fmt.Sprintf("candidate event %d %d\n", rn.candidateElectionTs, rn.electionTimeoutMS))
	if _, ok := se.body.(TimerTick); ok { // Idle Timeout
		if IsTimeout(rn.candidateElectionTs, time.Now(), rn.electionTimeoutMS) {
			rn.print("reinit election")
			rn.switchToCandidate()

		}
	}
}

func (rn *RaftNode) candidateMsgEvent(message *MsgEvent) {
	if vr, ok := message.body.(VoteResponse); ok {
		vf := rn.getFollower(message.srcid)
		rn.print(fmt.Sprintf("vote response src=%d term=%d \n", message.srcid, vr.term))

		if vf != nil {
			vf.lastResponse = time.Now()
			vf.nextIndex = vr.lastLogIndex + 1
		}

		rn.VoteCount++
		fn := len(rn.followers)

		if rn.VoteCount > (fn+1)/2 {
			rn.switchToLeader()
		}
		return
	}

	if vr, ok := message.body.(VoteRequest); ok {
		rn.print(fmt.Sprintf("vote request src=%d term=%d \n", message.srcid, vr.term))
		if rn.CurrentTerm < vr.term {
			rn.switchToFollower()
			rn.CurrentTerm = vr.term

			if rn.commitedIndex <= vr.committedIndex {
				rn.grantVote(message.srcid, vr.term)
				rn.VotedFor = 0
			}
		}
	}

	if ar, ok := message.body.(AppendEntries); ok {
		if rn.CurrentTerm <= ar.term {
			rn.switchToFollower()
			rn.followerProcessEvent(message)
		}
	}

	if cmd, ok := message.body.(ClientCommand); ok {
		rn.send(msg(rn.id, message.srcid, &ClientCommendResponse{cmdId: cmd.id, success: false, leaderid: rn.leader.id}))
		return
	}

}

func (rn *RaftNode) candidateClientEvent(cm *ClientEvent) {
	fmt.Printf("Candate recieved client event")
}
