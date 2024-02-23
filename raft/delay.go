package nw

import "raft/src/raftApi"

type Delay struct {
	delay         int64 // delay in microseconds
	InputChannel  chan raftApi.MsgEvent
	OutputChannel chan raftApi.MsgEvent
}
