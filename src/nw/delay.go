package nw

import (
	"raft/src/raftApi"
	"time"
)

type Delay struct {
	IncomingChan chan raftApi.MsgEvent
	OutgoingChan chan raftApi.MsgEvent

	delay int64
}

func (d *Delay) GetIncomingChannel() chan raftApi.MsgEvent {
	return d.IncomingChan
}

func (d *Delay) SetIncomingChannel(c chan raftApi.MsgEvent) {
	d.IncomingChan = c
}

func (d *Delay) GetOutgoingChannel() chan raftApi.MsgEvent {
	return d.OutgoingChan
}

func (d *Delay) SetOutgoingChannel(c chan raftApi.MsgEvent) {
	d.OutgoingChan = c
}

func CreateDelay(delay int64) *Delay {

	incoming := make(chan raftApi.MsgEvent, CHANNELSIZE)
	outgoing := make(chan raftApi.MsgEvent, CHANNELSIZE)

	return &Delay{delay: delay, IncomingChan: incoming, OutgoingChan: outgoing}
}

func (d *Delay) Run() {
	for {
		x := <-d.IncomingChan
		msg := x

		current := time.Now()
		targetTs := msg.Ts.Add(time.Duration(d.delay) * time.Microsecond)
		sleepDuration := targetTs.Sub(current)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}

		d.OutgoingChan <- msg
	}
}
