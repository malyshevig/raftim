package client

import (
	"container/list"
	nw2 "raft/src/net"
	"raft/src/proto"
	"time"
)

func (cl *RaftClientNode) send(m *proto.MsgEvent) {
	cl.logger.Infof("client %s send msg %s", *cl, m.Body)
	cl.OutgoingChannel <- *m
	cl.lastMsgSentTs = time.Now()
}

func (cl *RaftClientNode) sendCommand(cmd *CmdProcess) {
	cl.send(nw2.Msg(cl.Id, cl.leaderId, proto.ClientCommand{Id: cmd.id, Cmd: cmd.cmd}))
}

func (cl *RaftClientNode) discoveLeader() {
	if cl.leaderId == 0 {
		cl.logger.Infof("client %s dicover leader", *cl)
		idx := cl.rand.Intn(len(cl.nodes))
		nodeId := cl.nodes[idx]
		cl.send(nw2.Msg(cl.Id, nodeId, proto.LeaderDiscoveryRequest{}))
	}
}

// reSendCommand use in case of leader changes
func (cl *RaftClientNode) reSendCommands() {
	cl.logger.Infof("%s reSendCommands %d", *cl, cl.cmdSentList.Len())
	if cl.leaderId > 0 {
		for c := cl.cmdSentList.Front(); c != nil; c = c.Next() {
			cmd := c.Value.(*CmdProcess)
			cl.sendCommand(cmd)
		}
	}
}

func (cl *RaftClientNode) sendCommands() {
	// need to get leader
	cl.logger.Infof("client %s sendCommands %v", *cl, cl.cmdList)
	if cl.leaderId > 0 {
		for c := cl.cmdList.Front(); c != nil; c = c.Next() {
			cmd := c.Value.(*CmdProcess)
			cl.sendCommand(cmd)
			cl.cmdSentList.PushBack(cmd)
		}
		cl.cmdList = list.List{}
	} else {
		cl.logger.Infof("client %s trying to send o unknown leader", *cl)

	}
}
