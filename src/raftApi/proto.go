package raftApi

import (
	"time"
)

type SystemEvent struct {
	Body interface{}
}

type MsgEvent struct {
	Srcid int
	Dstid int

	Ts   time.Time
	Body interface{}
}

// System events
type TimerTick struct {
}
