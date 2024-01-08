package raft

type TimerTick struct {
}

type Printable interface {
	toString()
}

type VoteRequest struct {
	srcId int
}

type VoteResponse struct {
	srcId int
}

type PingRequest struct {
	srcId int
}

type PingResponse struct {
	srcId int
}

type Event struct {
	Term int64

	Inner interface{}
}
