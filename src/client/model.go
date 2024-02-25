package client

import (
	"container/list"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"raft/src/net"
	"raft/src/proto"
	"raft/src/util"
	"time"
)

type Node struct {
	Id              int
	IncomingChannel chan proto.MsgEvent
	OutgoingChannel chan proto.MsgEvent
	ControlChannel  chan proto.SystemEvent
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

func (cl *RaftClientNode) GetIncomingChannel() chan proto.MsgEvent {
	return cl.IncomingChannel
}

func (cl *RaftClientNode) GetOutgoingChannel() chan proto.MsgEvent {
	return cl.OutgoingChannel
}

func (cl *RaftClientNode) SetIncomingChannel(c chan proto.MsgEvent) {
	cl.IncomingChannel = c
}

func (cl *RaftClientNode) SetOutgoingChannel(c chan proto.MsgEvent) {
	cl.OutgoingChannel = c
}

func (cl RaftClientNode) String() string {
	return fmt.Sprintf("client: %d leaderId: %d", cl.Id, cl.leaderId)
}

func NewClientNode(id int, config net.ClusterConfig) *RaftClientNode {
	leaderId := 0

	incomingChannel := make(chan proto.MsgEvent, net.CHANNELSIZE)
	outgoingChannel := make(chan proto.MsgEvent, net.CHANNELSIZE)
	clientChannel := make(chan CmdProcess, net.CHANNELSIZE)
	controlChannel := make(chan proto.SystemEvent, net.CHANNELSIZE)
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
