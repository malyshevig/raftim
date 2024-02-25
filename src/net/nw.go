package net

import (
	"fmt"
	"log"
	"os"
	"raft/src/proto"
	"time"
)

type ClusterConfig struct {
	Nodes []int
}

const (
	CHANNELSIZE = 1000000
	DELAY       = 100
)

func Msg(srcId int, dstId int, body interface{}) *proto.MsgEvent {
	return &proto.MsgEvent{Srcid: srcId, Dstid: dstId, Body: body, Ts: time.Now()}
}

func initTrace() {
	f, err := os.OpenFile("./trace.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
}

func trace(ev proto.MsgEvent) {
	f, err := os.OpenFile("./trace.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	s := fmt.Sprintf("%v: %d->%d: %s\n", ev.Ts, ev.Srcid, ev.Dstid, ev.Body)
	_, err = f.WriteString(s)
	if err != nil {
		log.Fatal(err)
	}
}

type NodeChainInf interface {
	GetIncomingChannel() chan proto.MsgEvent
	SetIncomingChannel(chan proto.MsgEvent)

	GetOutgoingChannel() chan proto.MsgEvent
	SetOutgoingChannel(chan proto.MsgEvent)
}

func makeInChannelIfNil(node NodeChainInf) {
	if node.GetIncomingChannel() == nil {
		channel := make(chan proto.MsgEvent, 100000)
		node.SetIncomingChannel(channel)
	}
}

func makeOutChannelIfNil(node NodeChainInf) {
	if node.GetOutgoingChannel() == nil {
		channel := make(chan proto.MsgEvent, 100000)
		node.SetOutgoingChannel(channel)
	}
}

func BuildNodesChain(nodes ...NodeChainInf) NodeChainInf {
	if len(nodes) == 0 {
		return nil
	}
	n0 := nodes[0]

	prevNode := n0
	nodes = nodes[1:]

	makeInChannelIfNil(n0)

	for _, n := range nodes {
		makeInChannelIfNil(n)

		prevNode.SetOutgoingChannel(n.GetIncomingChannel())
		prevNode = n
	}
	makeOutChannelIfNil(nodes[len(nodes)-1])
	return n0
}
