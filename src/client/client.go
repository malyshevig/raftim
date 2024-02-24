package client

import (
	"container/list"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	nw2 "raft/src/nw"
	"raft/src/raftApi"
	"raft/src/util"
	"sync"
	"sync/atomic"
	"time"
)

const InactiveTimeout = 1000

type Node struct {
	Id              int
	IncomingChannel chan raftApi.MsgEvent
	OutgoingChannel chan raftApi.MsgEvent
	ControlChannel  chan raftApi.SystemEvent
}

type RaftClientNode struct {
	Node
	ClientChannel chan CmdProcess

	leaderId        int
	clientCommandId int64

	cmdList     list.List
	cmdSentList list.List

	lastMsgReceivedTs time.Time
	lastMsgSentTs     time.Time

	nodes []int
	rand  *rand.Rand

	logger *zap.SugaredLogger
}

// Implementation of NodeChainInf

func (cl *RaftClientNode) GetIncomingChannel() chan raftApi.MsgEvent {
	return cl.IncomingChannel
}

func (cl *RaftClientNode) GetOutgoingChannel() chan raftApi.MsgEvent {
	return cl.OutgoingChannel
}

func (cl *RaftClientNode) SetIncomingChannel(c chan raftApi.MsgEvent) {
	cl.IncomingChannel = c
}

func (cl *RaftClientNode) SetOutgoingChannel(c chan raftApi.MsgEvent) {
	cl.OutgoingChannel = c
}

func (cl RaftClientNode) String() string {
	return fmt.Sprintf("client: %d leaderId: %d", cl.Id, cl.leaderId)
}

func NewClientNode(id int, config nw2.ClusterConfig) *RaftClientNode {
	leaderId := 0

	incomingChannel := make(chan raftApi.MsgEvent, nw2.CHANNELSIZE)
	outgoingChannel := make(chan raftApi.MsgEvent, nw2.CHANNELSIZE)
	clientChannel := make(chan CmdProcess, nw2.CHANNELSIZE)
	controlChannel := make(chan raftApi.SystemEvent, nw2.CHANNELSIZE)
	clientCommandId := int64(1)

	cmdList := list.List{}
	cmdSentList := list.List{}

	s := rand.NewSource(time.Now().Unix())
	rnd := rand.New(s)

	loggerFile := fmt.Sprintf("logs/client.log")
	logger := util.InitLogger(loggerFile)

	return &RaftClientNode{Node: Node{Id: id, IncomingChannel: incomingChannel,
		OutgoingChannel: outgoingChannel, ControlChannel: controlChannel},
		leaderId:      leaderId,
		ClientChannel: clientChannel, clientCommandId: clientCommandId,
		cmdList: cmdList, cmdSentList: cmdSentList,
		lastMsgReceivedTs: time.Now(), lastMsgSentTs: time.Now(),
		nodes:  config.Nodes,
		rand:   rnd,
		logger: logger.Sugar()}
}

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

func (cl *RaftClientNode) allocateCmdId() int64 {
	for {
		v := cl.clientCommandId
		if atomic.CompareAndSwapInt64(&cl.clientCommandId, v, v+1) {
			return v + 1
		}
	}
}

type CmdProcess struct {
	wg *sync.WaitGroup

	id    int64
	cmd   string
	state string
}

func (c CmdProcess) String() string {
	return fmt.Sprintf("%s", c.cmd)
}

type EventLeaderChanged struct {
	oldLeaderId int
	newLeaderId int
}

func (cl *RaftClientNode) commitCommands(cmdid int64) {
	cl.logger.Infof("%s commitCommands %d", *cl, cmdid)
	fmt.Printf("%v Client commitCommands %d\n", time.Now(), cmdid)
	for c := cl.cmdSentList.Front(); c != nil; c = c.Next() {
		cmd := c.Value.(*CmdProcess)
		cmd.state = "Done"
		cmd.wg.Done()

		cl.cmdSentList.Remove(c)
		if cmd.id == cmdid {
			break
		}
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

func (cl *RaftClientNode) processMsg(msg *raftApi.MsgEvent) {
	if resp, ok := msg.Body.(raftApi.ClientCommandResponse); ok {
		cl.lastMsgReceivedTs = time.Now()
		if resp.Success {
			cl.commitCommands(resp.CmdId)
		} else {
			oldLeaderId := cl.leaderId
			cl.ControlChannel <- raftApi.SystemEvent{Body: EventLeaderChanged{oldLeaderId: oldLeaderId, newLeaderId: resp.Leaderid}}
		}
		return
	}

	if resp, ok := msg.Body.(raftApi.LeaderDiscoveryResponse); ok {
		cl.lastMsgReceivedTs = time.Now()
		cl.logger.Infof("%s received LeaderDoscoveryResponse", *cl)

		if cl.leaderId != resp.LeaderId {
			oldLeaderId := cl.leaderId
			cl.ControlChannel <- raftApi.SystemEvent{Body: EventLeaderChanged{oldLeaderId: oldLeaderId, newLeaderId: resp.LeaderId}}
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

func (cl *RaftClientNode) sendCommand(cmd *CmdProcess) {
	cl.send(nw2.Msg(cl.Id, cl.leaderId, raftApi.ClientCommand{Id: cmd.id, Cmd: cmd.cmd}))
}

func (cl *RaftClientNode) processClientCmd(cmd *CmdProcess) {
	cl.cmdList.PushBack(cmd)
	cl.sendCommands()
}

func (cl *RaftClientNode) processEvent(e *raftApi.SystemEvent) {
	if leaderChanged, ok := e.Body.(EventLeaderChanged); ok {
		cl.logger.Infof("client %s changed leader oldleader: %d newLeader: %d ", *cl, leaderChanged.oldLeaderId, leaderChanged.newLeaderId)
		cl.leaderId = leaderChanged.newLeaderId
		if cl.leaderId > 0 {
			cl.reSendCommands()
			cl.sendCommands()
		}

		return
	}

	if _, ok := e.Body.(raftApi.TimerTick); ok {
		// regular check timeouts
		if nw2.IsTimeout(cl.lastMsgReceivedTs, time.Now(), InactiveTimeout) {
			cl.logger.Infof("client %s Inactive timeout  lastMsgReceivedTs %v, now: %v", *cl, cl.lastMsgReceivedTs, time.Now())

			cl.leaderId = 0
			cl.discoveLeader()
		}
		if nw2.IsTimeout(cl.lastMsgSentTs, time.Now(), InactiveTimeout/3) {
			// if no commands ping
			if cl.leaderId > 0 {
				cl.send(nw2.Msg(cl.Id, cl.leaderId, raftApi.LeaderDiscoveryRequest{}))
			}
		}

		return
	}

}

func (cl *RaftClientNode) send(m *raftApi.MsgEvent) {
	cl.logger.Infof("client %s send msg %s", *cl, m.Body)
	cl.OutgoingChannel <- *m
	cl.lastMsgSentTs = time.Now()
}

func (cl *RaftClientNode) discoveLeader() {
	if cl.leaderId == 0 {
		cl.logger.Infof("client %s dicover leader", *cl)
		idx := cl.rand.Intn(len(cl.nodes))
		nodeId := cl.nodes[idx]
		cl.send(nw2.Msg(cl.Id, nodeId, raftApi.LeaderDiscoveryRequest{}))
	}
}
