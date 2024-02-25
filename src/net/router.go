package net

import (
	"fmt"
	"raft/src/proto"
)

type Router struct {
	IncomingChan chan proto.MsgEvent
	routes       map[int]chan proto.MsgEvent
}

func (d *Router) GetIncomingChannel() chan proto.MsgEvent {
	return d.IncomingChan
}

func (d *Router) SetIncomingChannel(c chan proto.MsgEvent) {
	d.IncomingChan = c
}

func (d *Router) GetOutgoingChannel() chan proto.MsgEvent {
	return nil
}

func (d *Router) SetOutgoingChannel(chan proto.MsgEvent) {

}

func CreateRouter(incomingChan chan proto.MsgEvent) *Router {
	if incomingChan == nil {
		incomingChan = make(chan proto.MsgEvent, 1000000)
	}

	return &Router{IncomingChan: incomingChan, routes: make(map[int]chan proto.MsgEvent)}
}

func (r *Router) AddRoute(id int, channel chan proto.MsgEvent) {
	r.routes[id] = channel
}

func (r *Router) Run() {
	initTrace()
	for {
		x := <-r.IncomingChan
		msg := x
		trace(msg)
		ch := r.routes[msg.Dstid]

		if ch != nil {
			channel := ch
			channel <- x
		} else {
			fmt.Printf("channel not found for %d", msg.Dstid)
		}
	}
}
