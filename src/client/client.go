package client

import (
	"raft/src/net"
	"raft/src/proto"
	"reflect"
	"sync"
	"time"
)

const InactiveTimeout = 1000

func (cl RaftClientNode) ProcessRequest(cmd string) error {
	cmdProc := CmdProcess{wg: &sync.WaitGroup{}, id: cl.allocateCmdId(), cmd: cmd}

	cmdProc.wg.Add(1)
	cl.ClientChannel <- cmdProc
	cmdProc.wg.Wait()
	return nil
}

func (cl RaftClientNode) Run() {
	cl.discoveLeader()
	for {
		if cl.leaderId > 0 {
			select {
			case event := <-cl.ControlChannel:
				cl.processEvent(&event)
			case msg := <-cl.IncomingChannel:
				msgEvent := msg
				cl.processMsg(&msgEvent)

			case cmd := <-cl.ClientChannel:
				cl.processClientCmd(&cmd)
			}
		} else {
			select {
			case event := <-cl.ControlChannel:
				cl.processEvent(&event)
			case msg := <-cl.IncomingChannel:
				msgEvent := msg
				cl.processMsg(&msgEvent)
			}
		}
	}
}

func (cl *RaftClientNode) processMsg(msg *proto.MsgEvent) {
	if resp, ok := msg.Body.(proto.ClientCommandResponse); ok {
		cl.lastMsgReceivedTs = time.Now()
		if resp.Success {
			cl.commitCommands(resp.CmdId)
		} else {
			oldLeaderId := cl.leaderId
			cl.ControlChannel <- proto.SystemEvent{Body: EventLeaderChanged{oldLeaderId: oldLeaderId, newLeaderId: resp.Leaderid}}
		}
		return
	}

	if resp, ok := msg.Body.(proto.LeaderDiscoveryResponse); ok {
		cl.lastMsgReceivedTs = time.Now()
		cl.logger.Infof("%s received LeaderDoscoveryResponse", *cl)

		if cl.leaderId != resp.LeaderId {
			oldLeaderId := cl.leaderId
			cl.ControlChannel <- proto.SystemEvent{Body: EventLeaderChanged{oldLeaderId: oldLeaderId, newLeaderId: resp.LeaderId}}
		}
	}
}

func (cl *RaftClientNode) processEvent(e *proto.SystemEvent) {
	if leaderChanged, ok := e.Body.(EventLeaderChanged); ok {
		cl.logger.Infof("client %s changed leader oldleader: %d newLeader: %d ", *cl, leaderChanged.oldLeaderId, leaderChanged.newLeaderId)
		cl.leaderId = leaderChanged.newLeaderId
		if cl.leaderId > 0 {
			cl.reSendCommands()
			cl.sendCommands()
		}

		return
	}

	if _, ok := e.Body.(proto.TimerTick); ok {
		// regular check timeouts
		if net.IsTimeout(cl.lastMsgReceivedTs, time.Now(), InactiveTimeout) {
			cl.logger.Infof("client %s Inactive timeout  lastMsgReceivedTs %v, now: %v", *cl, cl.lastMsgReceivedTs, time.Now())

			cl.leaderId = 0
			cl.discoveLeader()
		}
		if net.IsTimeout(cl.lastMsgSentTs, time.Now(), InactiveTimeout/3) {
			// if no commands ping
			if cl.leaderId > 0 {
				cl.send(net.Msg(cl.Id, cl.leaderId, proto.LeaderDiscoveryRequest{}))
			}
		}

		return
	}
	cl.logger.Infof("client %s Unknown type of Event %s", *cl, reflect.TypeOf(e.Body))

}
