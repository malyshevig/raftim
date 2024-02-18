package raft

import (
	"fmt"
	"raft/nw"
	"raft/raftApi"
	"reflect"
	"time"
)

func (rn *RaftNode) followerProcessEvent(ev any) {
	if se, ok := ev.(raftApi.SystemEvent); ok {
		rn.followerProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(raftApi.MsgEvent); ok {
		rn.followerProcessMsgEvent(&msg)
		return
	}

	rn.logger.Infow("")
	fmt.Printf("unexpected Event type %s", reflect.TypeOf(ev))

	return
}

func (rn *RaftNode) followerProcessSystemEvent(ev *raftApi.SystemEvent) {
	if _, ok := ev.Body.(raftApi.TimerTick); ok { // Idle Timeout
		if nw.IsTimeout(rn.followerLeaderIdleTs, time.Now(), rn.FollowerTimeoutMS) {
			rn.print(fmt.Sprintf("switch to candidate followerLeaderIdleTs = %v\n", rn.followerLeaderIdleTs))
			rn.switchToCandidate()
			return
		}
	}
}

func (rn *RaftNode) followerProcessMsgEvent(message *raftApi.MsgEvent) {

	if vr, ok := message.Body.(raftApi.VoteRequest); ok {
		rn.followerProcessVoteRequest(message, vr)
		return
	}

	if ar, ok := message.Body.(raftApi.AppendEntries); ok {
		rn.followerProcessAE(message, ar)
		return
	}

	if cmd, ok := message.Body.(raftApi.ClientCommand); ok {
		rn.Send(nw.Msg(rn.Id, message.Srcid, raftApi.ClientCommandResponse{CmdId: cmd.Id, Success: false, Leaderid: rn.leader.id}))
		return
	}
	if _, ok := message.Body.(raftApi.LeaderDiscoveryRequest); ok {
		rn.Send(nw.Msg(rn.Id, message.Srcid, raftApi.LeaderDiscoveryResponse{LeaderId: rn.leader.id}))
		return
	}
	rn.print(fmt.Sprintf("unexpected Msg %v\n", reflect.TypeOf(message.Body)))

}

func (rn *RaftNode) followerProcessAE(msg *raftApi.MsgEvent, ar raftApi.AppendEntries) {
	if rn.CurrentTerm > ar.Term {
		rn.sendAEResponse(ar.Id, msg.Srcid, false)
		return
	}
	rn.print(fmt.Sprintf("follower recieved %v\n", ar))

	rn.followerLeaderIdleTs = time.Now()
	rn.leader.id = msg.Srcid
	rn.leader.leaderLastTS = time.Now()

	if !rn.checkLog(ar.LastLogIndex, ar.LastLogTerm) {
		rn.sendAEResponse(ar.Id, msg.Srcid, false)
		return
	}

	startIndex := ar.LastLogIndex + 1
	for _, entry := range ar.Entries {
		rn.appendLogEntry(startIndex, entry)
		rn.print(fmt.Sprintf("follower save %s\n", entry.Cmd))
		startIndex++
	}

	if ar.LeaderCommittedIndex > rn.CommitedIndex {
		rn.saveLog(rn.CommitedIndex+1, ar.LeaderCommittedIndex)
		rn.CommitedIndex = ar.LeaderCommittedIndex
		rn.print(fmt.Sprintf("follower shift CommitedIndex to %d", rn.CommitedIndex))
	}

	rn.sendAEResponse(ar.Id, msg.Srcid, true)
	return
}

func (rn *RaftNode) followerProcessVoteRequest(msg *raftApi.MsgEvent, vr raftApi.VoteRequest) {
	if rn.CurrentTerm < vr.Term {
		rn.CurrentTerm = vr.Term
		rn.switchToFollower(0)
		rn.VotedFor = 0

		if rn.CommitedIndex <= vr.CommittedIndex {
			rn.grantVote(msg.Srcid, rn.CurrentTerm)
			rn.VotedFor = msg.Srcid
			rn.followerLeaderIdleTs = time.Now()
		}
	} else {
		if rn.CurrentTerm == vr.Term && (msg.Srcid == rn.VotedFor || rn.VotedFor == 0) {
			rn.followerLeaderIdleTs = time.Now()
			if rn.VotedFor == 0 {
				if rn.CommitedIndex <= vr.CommittedIndex {
					rn.grantVote(msg.Srcid, rn.CurrentTerm)
					rn.VotedFor = msg.Srcid
				}
			}
		}
	}

	return
}

func (rn *RaftNode) sendAEResponse(ae_id int, leaderId int, success bool) {
	rn.print(fmt.Sprintf("follower Send ae response cmdLod=%d\n", len(rn.CmdLog)))
	m := nw.Msg(rn.Id, leaderId, raftApi.AppendEntriesResponse{Ae_id: ae_id, Success: success, LastIndex: len(rn.CmdLog) - 1})
	rn.Send(m)
}
