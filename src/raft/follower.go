package raft

import (
	nw2 "raft/src/net"
	"raft/src/proto"
	"time"
)

func (rn *RaftNode) followerProcessEvent(ev any) {
	if se, ok := ev.(proto.SystemEvent); ok {
		rn.followerProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(proto.MsgEvent); ok {
		rn.followerProcessMsgEvent(&msg)
		return
	}

	rn.logger.Infow("%s unexpected Event type ", *rn)

	return
}

func (rn *RaftNode) followerProcessSystemEvent(ev *proto.SystemEvent) {
	if _, ok := ev.Body.(proto.TimerTick); ok { // Idle Timeout
		if nw2.IsTimeout(rn.followerLeaderIdleTs, time.Now(), rn.FollowerTimeoutMS) {
			rn.logger.Infof("%s switch to candidate followerLeaderIdleTs = %v", *rn, rn.followerLeaderIdleTs)

			rn.switchToCandidate()
			return
		}
	}
}

func (rn *RaftNode) followerProcessMsgEvent(message *proto.MsgEvent) {
	//rn.logger.Infof("%s follower process msg ", *rn)

	if vr, ok := message.Body.(proto.VoteRequest); ok {
		rn.followerProcessVoteRequest(message, vr)
		return
	}

	if ar, ok := message.Body.(proto.AppendEntries); ok {
		rn.followerProcessAE(message, ar)
		return
	}

	if cmd, ok := message.Body.(proto.ClientCommand); ok {
		rn.Send(nw2.Msg(rn.Id, message.Srcid, proto.ClientCommandResponse{CmdId: cmd.Id, Success: false, Leaderid: rn.leader.id}))
		return
	}
	if _, ok := message.Body.(proto.LeaderDiscoveryRequest); ok {
		rn.Send(nw2.Msg(rn.Id, message.Srcid, proto.LeaderDiscoveryResponse{LeaderId: rn.leader.id}))
		return
	}
	rn.logger.Infof("%s unexpected message type", *rn)

}

func (rn *RaftNode) followerProcessAE(msg *proto.MsgEvent, ar proto.AppendEntries) {
	if rn.CurrentTerm > ar.Term {
		rn.sendAEResponse(ar.Id, msg.Srcid, false)
		return
	}
	rn.logger.Infof("%s follower process ae %s", *rn, ar)

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

		rn.logger.Infof("%s follower save cmd:%s", *rn, entry.Cmd)
		startIndex++
	}

	if ar.LeaderCommittedIndex > rn.CommitedIndex {
		rn.saveLog(rn.CommitedIndex+1, ar.LeaderCommittedIndex)
		rn.CommitedIndex = ar.LeaderCommittedIndex
		rn.logger.Infof("%s follower shift CommitedIndex to %d", *rn, rn.CommitedIndex)

	}

	rn.sendAEResponse(ar.Id, msg.Srcid, true)
	return
}

func (rn *RaftNode) followerProcessVoteRequest(msg *proto.MsgEvent, vr proto.VoteRequest) {
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
	m := nw2.Msg(rn.Id, leaderId, proto.AppendEntriesResponse{AeId: ae_id, Success: success, LastIndex: len(rn.CmdLog) - 1})
	rn.Send(m)
}
