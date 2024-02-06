package raft

import (
	"time"
)

type Delay struct {
	delay         int64 // delay in microseconds
	inputChannel  chan interface{}
	outputChannel chan interface{}
}

type Router struct {
	incomingChannel chan interface{}
	routes          map[int]*chan interface{}
}

func (rn *RaftNode) send(msg interface{}) error {

	ch := rn.Node.OutgoingChan
	ch <- msg
	return nil
}

func CreateDelay(delay int64, incoming chan interface{}, outgoing chan interface{}) *Delay {
	if incoming == nil {
		incoming = make(chan interface{}, CHANNELSIZE)
	}
	if outgoing == nil {
		outgoing = make(chan interface{}, CHANNELSIZE)
	}

	return &Delay{delay: delay, inputChannel: incoming, outputChannel: outgoing}
}

func (d *Delay) run() {
	for {
		x := <-d.inputChannel
		msg := x.(MsgEvent)

		current := time.Now()
		targetTs := msg.ts.Add(time.Duration(d.delay) * time.Millisecond)
		sleepDuration := targetTs.Sub(current)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}

		d.outputChannel <- msg
	}
}

func CreateRouter(incomingChan chan interface{}) *Router {
	if incomingChan == nil {
		incomingChan = make(chan interface{}, 1000000)
	}

	return &Router{incomingChannel: incomingChan, routes: make(map[int]*chan interface{})}
}

func (r *Router) AddRoute(id int, channel chan interface{}) {
	r.routes[id] = &channel
}

func (r *Router) run() {
	initTrace()
	for {
		x := <-r.incomingChannel
		msg := x.(MsgEvent)
		trace(msg)
		channel := *r.routes[msg.dstid]

		channel <- x
	}
}

func msg(srcId int, dstId int, body interface{}) MsgEvent {
	return MsgEvent{srcid: srcId, dstid: dstId, body: body, ts: time.Now()}
}
