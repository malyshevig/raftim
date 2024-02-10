package raft

import (
	"fmt"
	"reflect"
	"time"
)

func (rn *RaftNode) leaderProcessEvent(ev interface{}) {
	// process VoteRequest

	if se, ok := ev.(SystemEvent); ok {
		rn.leaderProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(MsgEvent); ok {
		rn.leaderProcessMsgEvent(&msg)
		return
	}

	//if cm, ok := ev.(ClientEvent); ok {
	//	rn.leaderClientEvent(&cm)
	//	return
	//}

	rs := fmt.Sprintf("unexpected Event type %s", reflect.TypeOf(ev))
	rn.print(rs)
	return
}

func (rn *RaftNode) leaderProcessSystemEvent(ev *SystemEvent) {
	if _, ok := ev.body.(TimerTick); ok { // Idle Timeout
		if IsTimeout(rn.leaderPingTs, time.Now(), LeaderPingInterval) {
			rn.print("Ping ALl")
			rn.leaderPingTs = time.Now()
			rn.sendPingToAll()
		}
	}
}

func (rn *RaftNode) sendPingToAll() {
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
	if newCommittedIndex > rn.commitedIndex {
		rn.saveLog(rn.commitedIndex+1, newCommittedIndex)
		rn.commitedIndex = newCommittedIndex
	}

	rn.syncFollowers(false)
}

func (rn *RaftNode) ackCommands(from int, to int) {
	for idx := from; idx <= to; idx++ {
		cmd := rn.CmdLog[idx]
		msg := msg(rn.id, cmd.clientId, ClientCommendResponse{cmdId: cmd.msgId, success: true})
		rn.send(msg)
	}

}

func (rn *RaftNode) syncFollowers(delay bool) {
	rn.print("Sync followers")
	for _, fv := range rn.followers {
		remote := ClusterInstance().getNode(fv.id)

		rn.print(fmt.Sprintf("fv.id = %d fv.next=%d leader.CmdLog=%d leader.commited= %d", fv.id, fv.nextIndex, len(rn.CmdLog), rn.commitedIndex))
		rn.print(fmt.Sprintf("remote.id = %d remote.CmdLog=%d remote.committed=%d", remote.id, len(remote.CmdLog), remote.commitedIndex))

		if delay && !IsTimeout(fv.lastRequest, time.Now(), 100) {
			continue
		}

		entriesToSend := []Entry{}
		if fv.nextIndex < len(rn.CmdLog) {
			entriesToSend = rn.CmdLog[fv.nextIndex:]
		}

		prevIndex := fv.nextIndex - 1

		var prevTerm int64
		if prevIndex >= 0 {
			if prevIndex >= len(rn.CmdLog) {
				rn.print(fmt.Sprintf("prevIndex %d > rn.CmdLog() %d ", prevIndex, len(rn.CmdLog)))
			}

			prevTerm = rn.CmdLog[prevIndex].term
		} else {
			prevTerm = 0
		}

		//rs := fmt.Sprintf("sync %d num_entries= %d  index = %d  %v\n", fv.id, len(entriesToSend), fv.nextIndex)
		//rn.print(rs)
		event := msg(rn.id, fv.id, AppendEntries{id: rn.ae_id, entries: entriesToSend, lastLogIndex: fv.nextIndex - 1, term: rn.CurrentTerm,
			lastLogTerm: prevTerm, leaderCommittedIndex: rn.commitedIndex})
		rn.ae_id++
		rn.send(event)

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

func (rn *RaftNode) leaderProcessMsgEvent(ev *MsgEvent) {

	if vr, ok := ev.body.(VoteRequest); ok {
		if rn.CurrentTerm < vr.term {
			rn.CurrentTerm = vr.term
			rn.switchToFollower()
			rn.VotedFor = 0

			if rn.commitedIndex <= vr.committedIndex {
				rn.grantVote(ev.srcid, vr.term)
				rn.VotedFor = ev.srcid
			}
		}
		return
	}

	if ae, ok := ev.body.(AppendEntriesResponse); ok {
		rn.leaderProcessAEResponse(ev, ae)
		return
	}

	if _, ok := ev.body.(VoteResponse); ok {
		return
	}

	if cmd, ok := ev.body.(ClientCommand); ok {
		rn.appendLog(cmd.id, ev.srcid, cmd.cmd)
		return
	}

	fmt.Printf("Unexpected msg type recieved %s\n", reflect.TypeOf(ev.body))
}

func (rn *RaftNode) leaderProcessAEResponse(ev *MsgEvent, ae AppendEntriesResponse) bool {
	if !ae.success {
		return true
	}
	vf := rn.getFollower(ev.srcid)

	if vf != nil {

		vf.nextIndex = ae.lastIndex + 1
		vf.lastResponse = time.Now()

		//			rs := fmt.Sprintf("response vf = %d vf.NextIndex=%d", vf.id, vf.nextIndex)
		//			rn.print(rs)
		rn.commitInfo.updateFollowerIndex(vf)
		committedIndex := rn.commitInfo.GetNewCommitIndex()

		if committedIndex > rn.commitedIndex {
			rn.saveLog(rn.commitedIndex+1, committedIndex)
			rn.ackCommands(rn.commitedIndex+1, committedIndex)
			rn.commitedIndex = committedIndex
			rn.print(fmt.Sprintf("Leader shift commitedIndex to %d", rn.commitedIndex))
			rn.syncFollowers(true)
		}
	} else {
		rn.print("Follower not found")
	}
	return false
}

func (rn *RaftNode) appendLog(msgId int64, clientId int, cmd string) {
	rn.CmdLog = append(rn.CmdLog, Entry{term: rn.CurrentTerm, cmd: cmd, clientId: clientId, msgId: msgId})
	//logWriteStr(rn.id, cmd)
}

//func (rn *RaftNode) leaderClientEvent(ev *ClientEvent) {
//	if vr, ok := ev.body.(ClientCommand); ok {
//		rn.appendLog(vr.cmd)
//		rn.syncFollowers(true)
//
//		return
//	}
//}

func (rn *RaftNode) makeFollowers() map[int]*FollowerInfo {
	followers := make(map[int]*FollowerInfo, 0)
	for n := ClusterInstance().GetNodes().Front(); n != nil; n = n.Next() {
		raftNode := n.Value.(*RaftNode)
		if raftNode.id != rn.id {
			followers[raftNode.id] = &FollowerInfo{raftNode.id, time.Now(), time.Now(), 0}
		}
	}
	return followers
}
