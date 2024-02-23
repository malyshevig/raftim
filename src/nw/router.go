package nw

import (
	"fmt"
	"raft/src/raftApi"
)

type Router struct {
	IncomingChan chan raftApi.MsgEvent
	routes       map[int]chan raftApi.MsgEvent
}

func (d *Router) GetIncomingChannel() chan raftApi.MsgEvent {
	return d.IncomingChan
}

func (d *Router) SetIncomingChannel(c chan raftApi.MsgEvent) {
	d.IncomingChan = c
}

func (d *Router) GetOutgoingChannel() chan raftApi.MsgEvent {
	return nil
}

func (d *Router) SetOutgoingChannel(c chan raftApi.MsgEvent) {

}

func CreateRouter(incomingChan chan raftApi.MsgEvent) *Router {
	if incomingChan == nil {
		incomingChan = make(chan raftApi.MsgEvent, 1000000)
	}

	return &Router{IncomingChan: incomingChan, routes: make(map[int]chan raftApi.MsgEvent)}
}

func (r *Router) AddRoute(id int, channel chan raftApi.MsgEvent) {
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
