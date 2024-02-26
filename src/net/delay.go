package net

import (
	"raft/src/proto"
	"time"
)

type Delay struct {
	IncomingChan chan proto.MsgEvent
	OutgoingChan chan proto.MsgEvent

	delay int
}

func (d *Delay) GetIncomingChannel() chan proto.MsgEvent {
	return d.IncomingChan
}

func (d *Delay) SetIncomingChannel(c chan proto.MsgEvent) {
	d.IncomingChan = c
}

func (d *Delay) GetOutgoingChannel() chan proto.MsgEvent {
	return d.OutgoingChan
}

func (d *Delay) SetOutgoingChannel(c chan proto.MsgEvent) {
	d.OutgoingChan = c
}

func CreateDelay(delay int) *Delay {

	incoming := make(chan proto.MsgEvent, CHANNELSIZE)
	outgoing := make(chan proto.MsgEvent, CHANNELSIZE)

	return &Delay{delay: delay, IncomingChan: incoming, OutgoingChan: outgoing}
}

func (d *Delay) Run() {
	for {
		x := <-d.IncomingChan
		msg := x

		current := time.Now()

		targetTs := msg.Ts.Add(time.Duration(RandomiseTimeout(d.delay)) * time.Microsecond)
		sleepDuration := targetTs.Sub(current)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}

		d.OutgoingChan <- msg
	}
}
