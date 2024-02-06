package raft

import (
	"fmt"
	"reflect"
	"time"
)

func (rn *RaftNode) followerProcessEvent(ev any) {
	if se, ok := ev.(SystemEvent); ok {
		rn.followerProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(MsgEvent); ok {
		rn.followerMsgEvent(&msg)
		return
	}

	if cm, ok := ev.(ClientEvent); ok {
		rn.followerClientEvent(&cm)
		return
	}

	fmt.Printf("unexpected Event type %s", reflect.TypeOf(ev))
	return
}

func (rn *RaftNode) followerProcessSystemEvent(ev *SystemEvent) {
	if _, ok := ev.body.(TimerTick); ok { // Idle Timeout
		if IsTimeout(rn.followerLeaderIdleTs, time.Now(), rn.followerTimeoutMS) {
			rn.switchToCandidate()
			return
		}
	}
}

func (rn *RaftNode) followerMsgEvent(msg *MsgEvent) {

	if vr, ok := msg.body.(VoteRequest); ok {
		rn.followerProcessVoteRequest(msg, vr)
		return
	}

	if ar, ok := msg.body.(AppendEntries); ok {
		rn.followerProcessAE(msg, ar)
		return
	}

	if _, ok := msg.body.(ClientCommand); ok {
		//rn.print("follower received client command %s", vr.cmd)
		return
	}

	fmt.Printf("unexpected Msg %s\n", reflect.TypeOf(msg.body))

}

func (rn *RaftNode) followerProcessAE(msg *MsgEvent, ar AppendEntries) {
	if rn.CurrentTerm > ar.term {
		rn.sendAEResponse(ar.id, msg.srcid, false)
		return
	}

	rn.followerLeaderIdleTs = time.Now()
	if ar.leaderCommittedIndex > rn.commitedIndex {
		rn.saveLog(rn.commitedIndex+1, ar.leaderCommittedIndex)
		rn.commitedIndex = ar.leaderCommittedIndex
		rn.print(fmt.Sprintf("follower shift commitedIndex to %d", rn.commitedIndex))
	}

	if !rn.checkLog(ar.lastLogIndex, ar.lastLogTerm) {
		rn.sendAEResponse(ar.id, msg.srcid, false)
		return
	}

	startIndex := ar.lastLogIndex + 1
	for _, entry := range ar.entries {
		rn.appendLogEntry(startIndex, entry)
		rn.print(fmt.Sprintf("follower save %s\n", entry.cmd))
		startIndex++
	}
	rn.sendAEResponse(ar.id, msg.srcid, true)
	return
}

func (rn *RaftNode) followerProcessVoteRequest(msg *MsgEvent, vr VoteRequest) {
	if rn.CurrentTerm < vr.term {
		rn.followerLeaderIdleTs = time.Now()
		rn.CurrentTerm = vr.term
		rn.switchToFollower()
		rn.VotedFor = 0

		if rn.commitedIndex <= vr.committedIndex {
			rn.grantVote(msg.srcid, rn.CurrentTerm)
			rn.VotedFor = msg.srcid
		}

	} else {
		if rn.CurrentTerm == vr.term && (msg.srcid == rn.VotedFor || rn.VotedFor == 0) {
			rn.followerLeaderIdleTs = time.Now()
			rn.VotedFor = 0
			if rn.commitedIndex <= vr.committedIndex {
				rn.grantVote(msg.srcid, rn.CurrentTerm)
				rn.VotedFor = msg.srcid
			}
		}
	}

	return
}

func (rn *RaftNode) followerClientEvent(cm *ClientEvent) {
	fmt.Printf("Follower recieved client event")
}

func (rn *RaftNode) sendAEResponse(ae_id int, leaderId int, success bool) {
	cluster := ClusterInstance()
	leader := cluster.getNode(leaderId)
	if len(leader.CmdLog) <= len(rn.CmdLog)-1 {
		rn.print("logs mismatch")
	}

	rn.print(fmt.Sprintf("follower send ae response cmdLod=%d\n", len(rn.CmdLog)))
	m := msg(rn.Id, leaderId, AppendEntriesResponse{ae_id: ae_id, success: success, lastIndex: len(rn.CmdLog) - 1})
	rn.send(m)
}