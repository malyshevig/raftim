package net

import (
	"math/rand"
	"raft/src/proto"
	"time"
)

type TickGenerator struct {
	channels []chan proto.SystemEvent
}

func NewTickGenerator(channels []chan proto.SystemEvent) *TickGenerator {
	return &TickGenerator{channels: channels}
}

func (t *TickGenerator) Register(ch chan proto.SystemEvent) {
	t.channels = append(t.channels, ch)
}

func (t *TickGenerator) Run(tmMs int) {
	for {
		time.Sleep(time.Duration(int64(tmMs) * int64(time.Millisecond)))

		msg := proto.SystemEvent{Body: proto.TimerTick{}}
		for _, ch := range t.channels {
			ch <- msg
		}
	}
}

func IsTimeout(ts time.Time, now time.Time, intervalMS int) bool {
	targetTs := ts.Add(time.Duration(intervalMS) * time.Millisecond)

	return now.After(targetTs)
}

func InitRandomizer() *rand.Rand {
	src := rand.NewSource(time.Now().Unix())

	return rand.New(src)
}

var rnd = InitRandomizer()

func RandomiseTimeout(timeout int) int {
	div := timeout / 10

	return timeout + rnd.Intn(div) - div/2
}
