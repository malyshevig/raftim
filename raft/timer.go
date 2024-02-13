package raft

import "time"

type TickGenerator struct {
	channels []chan SystemEvent
}

func (t *TickGenerator) addChan(ch chan SystemEvent) {
	t.channels = append(t.channels, ch)
}

func (t *TickGenerator) run(tmMs int) {
	for {
		time.Sleep(time.Duration(int64(tmMs) * int64(time.Millisecond)))

		msg := SystemEvent{body: TimerTick{}}
		for _, ch := range t.channels {
			ch <- msg
		}
	}
}
