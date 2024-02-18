package raft

import (
	"fmt"
	"raft/nw"
	"raft/raftApi"
	"reflect"
	"time"
)

func (rn *RaftNode) leaderProcessEvent(ev interface{}) {
	if se, ok := ev.(raftApi.SystemEvent); ok {
		rn.leaderProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(raftApi.MsgEvent); ok {
		rn.leaderProcessMsgEvent(&msg)
		return
	}

	rs := fmt.Sprintf("unexpected Event type %s", reflect.TypeOf(ev))
	rn.print(rs)
	return
}

func (rn *RaftNode) leaderProcessSystemEvent(ev *raftApi.SystemEvent) {
	if _, ok := ev.Body.(raftApi.TimerTick); ok { // Idle Timeout
		if nw.IsTimeout(rn.leaderPingTs, time.Now(), LeaderPingInterval) {
			rn.leaderPingTs = time.Now()
			rn.sendPingToAll()
		}
	}
}

func (rn *RaftNode) sendPingToAll() {
	rn.logger.Infow("ping followers", "Node")
	rn.syncFollowers(true)
}

func (rn *RaftNode) switchToLeader() {
	rn.print("switch to Leader \n")

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
		msg := nw.Msg(rn.Id, cmd.ClientId, raftApi.ClientCommandResponse{CmdId: cmd.MsgId, Success: true})
		rn.Send(msg)
	}

}

func (rn *RaftNode) syncFollowers(delay bool) {
	rn.print("Sync followers")
	for _, fv := range rn.followers {

		if delay && !nw.IsTimeout(fv.lastRequest, time.Now(), 10) {
			continue
		}

		entriesToSend := []raftApi.Entry{}
		if fv.nextIndex < len(rn.CmdLog) {
			entriesToSend = rn.CmdLog[fv.nextIndex:]
		}

		prevIndex := fv.nextIndex - 1

		var prevTerm int64
		if prevIndex >= 0 {
			if prevIndex >= len(rn.CmdLog) {
				rn.print(fmt.Sprintf("prevIndex %d > rn.CmdLog() %d ", prevIndex, len(rn.CmdLog)))
			}

			prevTerm = rn.CmdLog[prevIndex].Term
		} else {
			prevTerm = 0
		}

		//rs := fmt.Sprintf("sync %d num_entries= %d  index = %d  %v\n", fv.Id, len(entriesToSend), fv.nextIndex)
		//rn.print(rs)
		event := nw.Msg(rn.Id, fv.id, raftApi.AppendEntries{Id: rn.ae_id, Entries: entriesToSend, LastLogIndex: fv.nextIndex - 1, Term: rn.CurrentTerm,
			LastLogTerm: prevTerm, LeaderCommittedIndex: rn.CommitedIndex})
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
			commit = min(commit, fv.nextIndex-1)
		}
	}
	return commit
}

func (rn *RaftNode) leaderProcessMsgEvent(msg *raftApi.MsgEvent) {

	if vr, ok := msg.Body.(raftApi.VoteRequest); ok {
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

	if ae, ok := msg.Body.(raftApi.AppendEntriesResponse); ok {
		rn.leaderProcessAEResponse(msg, ae)
		return
	}

	if _, ok := msg.Body.(raftApi.VoteResponse); ok {
		return
	}

	if cmd, ok := msg.Body.(raftApi.ClientCommand); ok {
		rn.print(fmt.Sprintf("Leader recieved clientCommand %d", cmd.Cmd))
		rn.appendLog(cmd.Id, msg.Srcid, cmd.Cmd)
		rn.syncFollowers(false)
		return
	}
	if _, ok := msg.Body.(raftApi.LeaderDiscoveryRequest); ok {
		rn.Send(nw.Msg(rn.Id, msg.Srcid, raftApi.LeaderDiscoveryResponse{LeaderId: rn.Id}))
		return
	}

	if ae, ok := msg.Body.(raftApi.AppendEntries); ok {
		if rn.CurrentTerm < ae.Term {
			rn.print(fmt.Sprintf("Leader recieved Append from the node %d with term = %d, switch to Follower", msg.Srcid, ae.Term))
			rn.CurrentTerm = ae.Term
			rn.switchToFollower(msg.Srcid)
		}
		return
	}

	fmt.Printf("Unexpected Msg type recieved %s\n", reflect.TypeOf(msg.Body))
}

func (rn *RaftNode) leaderProcessAEResponse(ev *raftApi.MsgEvent, ae raftApi.AppendEntriesResponse) bool {
	if !ae.Success {
		return true
	}
	vf := rn.getFollower(ev.Srcid)

	if vf != nil {

		vf.nextIndex = ae.LastIndex + 1
		vf.lastResponse = time.Now()

		//			rs := fmt.Sprintf("response vf = %d vf.NextIndex=%d", vf.Id, vf.nextIndex)
		//			rn.print(rs)
		rn.commitInfo.updateFollowerIndex(vf)
		committedIndex := rn.commitInfo.GetNewCommitIndex()

		if committedIndex > rn.CommitedIndex {
			rn.saveLog(rn.CommitedIndex+1, committedIndex)
			rn.ackCommands(rn.CommitedIndex+1, committedIndex)
			rn.CommitedIndex = committedIndex
			rn.print(fmt.Sprintf("Leader shift CommitedIndex to %d", rn.CommitedIndex))
			rn.syncFollowers(true)
		}
	} else {
		rn.print("Follower not found")
	}
	return false
}

func (rn *RaftNode) appendLog(msgId int64, clientId int, cmd string) {
	rn.CmdLog = append(rn.CmdLog, raftApi.Entry{Term: rn.CurrentTerm, Cmd: cmd, ClientId: clientId, MsgId: msgId})
	//logWriteStr(rn.Id, cmd)
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
