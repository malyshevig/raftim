package raft

import (
	net "raft/src/net"
	"raft/src/proto"
	"time"
)

func (rn *RaftNode) candidateProcessEvent(ev any) {

	if se, ok := ev.(proto.SystemEvent); ok {
		rn.candidateProcessSystemEvent(&se)
		return
	}

	if msg, ok := ev.(proto.MsgEvent); ok {
		rn.candidateProcessMsgEvent(&msg)
		return
	}

	rn.logger.Infof("%s unknown event %v", *rn, ev)
	return
}

func (rn *RaftNode) switchToCandidate() {
	rn.logger.Infof("%s switch to candidae", *rn)

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
		m := net.Msg(rn.Id, f.id, proto.VoteRequest{Term: rn.CurrentTerm, CommittedIndex: rn.CommitedIndex})
		rn.Send(m)
	}

}

func (rn *RaftNode) candidateProcessSystemEvent(se *proto.SystemEvent) {

	if _, ok := se.Body.(proto.TimerTick); ok { // Idle Timeout
		if net.IsTimeout(rn.candidateElectionTs, time.Now(), rn.ElectionTimeoutMS) {
			rn.logger.Infof("%s reInit Election process", *rn)
			rn.switchToCandidate()

		}
	}
}

func (rn *RaftNode) candidateProcessMsgEvent(message *proto.MsgEvent) {
	rn.logger.Infof("%s process msg %v", *rn, message)

	if vr, ok := message.Body.(proto.VoteResponse); ok {
		vf := rn.getFollower(message.Srcid)

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

	if vr, ok := message.Body.(proto.VoteRequest); ok {

		if rn.CurrentTerm < vr.Term {
			rn.switchToFollower(message.Srcid)
			rn.CurrentTerm = vr.Term

			if rn.CommitedIndex <= vr.CommittedIndex {
				rn.grantVote(message.Srcid, vr.Term)
				rn.VotedFor = 0

			}
		}
	}

	if ar, ok := message.Body.(proto.AppendEntries); ok {
		if rn.CurrentTerm <= ar.Term {
			rn.logger.Infof("%s recieved ae from %d with term %d", *rn, message.Srcid, ar.Term)
			rn.CurrentTerm = ar.Term

			rn.switchToFollower(message.Srcid)
			rn.followerProcessMsgEvent(message)
		}
		return
	}

	if cmd, ok := message.Body.(proto.ClientCommand); ok {
		rn.Send(net.Msg(rn.Id, message.Srcid, proto.ClientCommandResponse{CmdId: cmd.Id, Success: false, Leaderid: rn.leader.id}))
		return
	}

}
