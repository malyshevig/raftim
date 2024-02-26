package raft

import (
	"raft/src/net"
	"raft/src/proto"
	"raft/src/util"
	"time"
)

func (rn *RaftNode) leaderProcessEvent(ev interface{}) {
	if se, ok := ev.(proto.SystemEvent); ok {
		rn.leaderProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(proto.MsgEvent); ok {
		rn.leaderProcessMsgEvent(&msg)
		return
	}

	rn.logger.Infow("%s unexpected Event type ", *rn)
	return
}

func (rn *RaftNode) leaderProcessSystemEvent(ev *proto.SystemEvent) {
	if _, ok := ev.Body.(proto.TimerTick); ok { // Idle Timeout
		if net.IsTimeout(rn.leaderPingTs, time.Now(), LeaderPingInterval) {
			rn.leaderPingTs = time.Now()

			// Send ping to followers if needed
			rn.syncFollowers(true)
		}
	}
}

func (rn *RaftNode) switchToLeader() {
	rn.logger.Infof("%s switch to leader ", *rn)

	rn.State = Leader
	//	rn.followers = rn.makeFollowers()  - followes was made on  switching to candidate
	rn.commitInfo = InitCommittedInfo()
	for _, v := range rn.followers {
		rn.commitInfo.addFollower(v)
	}
	newCommittedIndex := rn.commitInfo.GetNewCommitIndex()
	if newCommittedIndex > rn.CommitedIndex {
		rn.saveLog(rn.CommitedIndex+1, newCommittedIndex)
		rn.CommitedIndex = newCommittedIndex
	}

	rn.syncFollowers(false)
}

func (rn *RaftNode) ackCommands(from int, to int) {
	for idx := from; idx <= to; idx++ {
		cmd := rn.CmdLog[idx]
		msg := net.Msg(rn.Id, cmd.ClientId, proto.ClientCommandResponse{CmdId: cmd.MsgId, Success: true})
		rn.Send(msg)
	}
}

func (rn *RaftNode) syncFollowers(delay bool) {
	rn.logger.Infow("%s sync followers ", *rn)

	for _, fv := range rn.followers {

		if delay && !net.IsTimeout(fv.lastRequest, time.Now(), 10) {
			continue
		}

		var entriesToSend []proto.Entry
		if fv.nextIndex < len(rn.CmdLog) {
			entriesToSend = rn.CmdLog[fv.nextIndex:]
		}

		prevIndex := fv.nextIndex - 1

		var prevTerm int64
		if prevIndex >= 0 {
			prevTerm = rn.CmdLog[prevIndex].Term
		} else {
			prevTerm = 0
		}

		event := net.Msg(rn.Id, fv.id,
			proto.AppendEntries{
				Id:                   rn.ae_id,
				Term:                 rn.CurrentTerm,
				Entries:              entriesToSend,
				LastLogIndex:         fv.nextIndex - 1,
				LastLogTerm:          prevTerm,
				LeaderCommittedIndex: rn.CommitedIndex})

		rn.ae_id++
		rn.Send(event)

		fv.lastRequest = time.Now()
	}
}

func (rn *RaftNode) leaderCalculateNewCommitIndex() int {
	commit := -1

	for _, fv := range rn.followers {
		if commit == -1 {
			commit = fv.nextIndex - 1
		} else {
			commit = util.Min(commit, fv.nextIndex-1)
		}
	}
	return commit
}

func (rn *RaftNode) leaderProcessMsgEvent(msg *proto.MsgEvent) {
	rn.logger.Infof("%s process message %s", *rn, msg)

	if vr, ok := msg.Body.(proto.VoteRequest); ok {
		if rn.CurrentTerm < vr.Term {
			rn.CurrentTerm = vr.Term
			rn.switchToFollower(0)
			rn.VotedFor = 0

			if rn.CommitedIndex <= vr.CommittedIndex {
				rn.grantVote(msg.Srcid, vr.Term)
				rn.VotedFor = msg.Srcid
			}
		}
		return
	}

	if ae, ok := msg.Body.(proto.AppendEntriesResponse); ok {
		rn.leaderProcessAEResponse(msg, ae)
		return
	}

	if _, ok := msg.Body.(proto.VoteResponse); ok {
		return
	}

	if cmd, ok := msg.Body.(proto.ClientCommand); ok {
		rn.appendLog(cmd.Id, msg.Srcid, cmd.Cmd)
		rn.syncFollowers(false)
		return
	}
	if _, ok := msg.Body.(proto.LeaderDiscoveryRequest); ok {
		rn.Send(net.Msg(rn.Id, msg.Srcid, proto.LeaderDiscoveryResponse{LeaderId: rn.Id}))
		return
	}

	if ae, ok := msg.Body.(proto.AppendEntries); ok {
		if rn.CurrentTerm < ae.Term {
			rn.logger.Infof("%s Leader recieved Append from the node %d with term = %d", *rn, msg.Srcid, ae.Term)

			rn.CurrentTerm = ae.Term
			rn.switchToFollower(msg.Srcid)
		}
		return
	}
	rn.logger.Infof("%s unexpected message type ", *rn)
}

func (rn *RaftNode) leaderProcessAEResponse(ev *proto.MsgEvent, ae proto.AppendEntriesResponse) bool {
	if !ae.Success {
		return true
	}
	vf := rn.getFollower(ev.Srcid)

	if vf != nil {

		vf.nextIndex = ae.LastIndex + 1
		vf.lastResponse = time.Now()

		rn.commitInfo.updateFollowerIndex(vf)
		committedIndex := rn.commitInfo.GetNewCommitIndex()

		if committedIndex > rn.CommitedIndex {
			rn.saveLog(rn.CommitedIndex+1, committedIndex)
			rn.ackCommands(rn.CommitedIndex+1, committedIndex)
			rn.CommitedIndex = committedIndex
			rn.logger.Infof("%s leader shift commit to %d", *rn, rn.CommitedIndex)

			rn.syncFollowers(true)
		}
	}
	return false
}

func (rn *RaftNode) appendLog(msgId int64, clientId int, cmd string) {
	rn.CmdLog = append(rn.CmdLog, proto.Entry{Term: rn.CurrentTerm, Cmd: cmd, ClientId: clientId, MsgId: msgId})
}

func (rn *RaftNode) makeFollowers() map[int]*FollowerInfo {
	followers := make(map[int]*FollowerInfo, 0)
	for _, nodeId := range rn.config.Nodes {
		if nodeId != rn.Id {
			followers[nodeId] = &FollowerInfo{nodeId, time.Now(), time.Now(), 0}
		}
	}
	return followers
}
