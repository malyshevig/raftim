package nw

import (
	"fmt"
	"log"
	"os"
	"raft/src/raftApi"
	"time"
)

type Delay struct {
	delay         int64 // delay in microseconds
	InputChannel  chan raftApi.MsgEvent
	OutputChannel chan raftApi.MsgEvent
}

type Router struct {
	IncomingChannel chan raftApi.MsgEvent
	routes          map[int]*chan raftApi.MsgEvent
}

const (
	CHANNELSIZE = 1000000
	DELAY       = 100
)

func CreateDelay(delay int64, incoming chan raftApi.MsgEvent, outgoing chan raftApi.MsgEvent) *Delay {
	if incoming == nil {
		incoming = make(chan raftApi.MsgEvent, CHANNELSIZE)
	}
	if outgoing == nil {
		outgoing = make(chan raftApi.MsgEvent, CHANNELSIZE)
	}

	return &Delay{delay: delay, InputChannel: incoming, OutputChannel: outgoing}
}

func (d *Delay) Run() {
	for {
		x := <-d.InputChannel
		msg := x

		current := time.Now()
		targetTs := msg.Ts.Add(time.Duration(d.delay) * time.Microsecond)
		sleepDuration := targetTs.Sub(current)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}

		d.OutputChannel <- msg
	}
}

func CreateRouter(incomingChan chan raftApi.MsgEvent) *Router {
	if incomingChan == nil {
		incomingChan = make(chan raftApi.MsgEvent, 1000000)
	}

	return &Router{IncomingChannel: incomingChan, routes: make(map[int]*chan raftApi.MsgEvent)}
}

func (r *Router) AddRoute(id int, channel chan raftApi.MsgEvent) {
	r.routes[id] = &channel
}

func (r *Router) Run() {
	initTrace()
	for {
		x := <-r.IncomingChannel
		msg := x
		trace(msg)
		ch := r.routes[msg.Dstid]

		if ch != nil {
			channel := *ch
			channel <- x
		} else {
			fmt.Printf("channel not found for %d", msg.Dstid)
		}
	}
}

func Msg(srcId int, dstId int, body interface{}) *raftApi.MsgEvent {
	return &raftApi.MsgEvent{Srcid: srcId, Dstid: dstId, Body: body, Ts: time.Now()}
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

func trace(ev raftApi.MsgEvent) {
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
	GetIncomingChannel() *chan raftApi.MsgEvent
	SetIncomingChannel(*chan raftApi.MsgEvent)

	GetOutgoingChannel() *chan raftApi.MsgEvent
	SetOutgoingChannel(*chan raftApi.MsgEvent)
}

func makeInChannelIfNil(node NodeChainInf) {
	if node.GetIncomingChannel() == nil {
		channel := make(chan raftApi.MsgEvent, 100000)
		node.SetIncomingChannel(&channel)
	}
}

func makeOutChannelIfNil(node NodeChainInf) {
	if node.GetOutgoingChannel() == nil {
		channel := make(chan raftApi.MsgEvent, 100000)
		node.SetOutgoingChannel(&channel)
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
