package raft

import (
	"fmt"
	"time"
)

type Delay struct {
	delay         int64 // delay in microseconds
	inputChannel  chan MsgEvent
	outputChannel chan MsgEvent
}

type Router struct {
	incomingChannel chan MsgEvent
	routes          map[int]*chan MsgEvent
}

func (rn *RaftNode) send(msg *MsgEvent) error {

	ch := rn.Node.outgoingChan
	ch <- *msg
	return nil
}

func CreateDelay(delay int64, incoming chan MsgEvent, outgoing chan MsgEvent) *Delay {
	if incoming == nil {
		incoming = make(chan MsgEvent, CHANNELSIZE)
	}
	if outgoing == nil {
		outgoing = make(chan MsgEvent, CHANNELSIZE)
	}

	return &Delay{delay: delay, inputChannel: incoming, outputChannel: outgoing}
}

func (d *Delay) run() {
	for {
		x := <-d.inputChannel
		msg := x

		current := time.Now()
		targetTs := msg.ts.Add(time.Duration(d.delay) * time.Microsecond)
		sleepDuration := targetTs.Sub(current)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}

		d.outputChannel <- msg
	}
}

func CreateRouter(incomingChan chan MsgEvent) *Router {
	if incomingChan == nil {
		incomingChan = make(chan MsgEvent, 1000000)
	}

	return &Router{incomingChannel: incomingChan, routes: make(map[int]*chan MsgEvent)}
}

func (r *Router) AddRoute(id int, channel chan MsgEvent) {
	r.routes[id] = &channel
}

func (r *Router) run() {
	initTrace()
	for {
		x := <-r.incomingChannel
		msg := x
		trace(msg)
		ch := r.routes[msg.dstid]

		if ch != nil {
			channel := *ch
			channel <- x
		} else {
			fmt.Printf("channel not found for %d", msg.dstid)
		}
	}
}

func msg(srcId int, dstId int, body interface{}) *MsgEvent {
	return &MsgEvent{srcid: srcId, dstid: dstId, body: body, ts: time.Now()}
}
