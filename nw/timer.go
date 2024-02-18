package nw

import (
	"math/rand"
	"raft/raftApi"
	"time"
)

type TickGenerator struct {
	channels []chan raftApi.SystemEvent
}

func NewTickGenerator(channels []chan raftApi.SystemEvent) *TickGenerator {
	return &TickGenerator{channels: channels}
}

func (t *TickGenerator) AddChan(ch chan raftApi.SystemEvent) {
	t.channels = append(t.channels, ch)
}

func (t *TickGenerator) Run(tmMs int) {
	for {
		time.Sleep(time.Duration(int64(tmMs) * int64(time.Millisecond)))

		msg := raftApi.SystemEvent{Body: raftApi.TimerTick{}}
		for _, ch := range t.channels {
			ch <- msg
		}
	}
}

func IsTimeout(ts time.Time, now time.Time, intervalMS int) bool {
	targetTs := ts.Add(time.Duration(intervalMS) * time.Millisecond)

	return now.After(targetTs)
}

type Rnd struct {
	r *rand.Rand
}

func InitRand() *Rnd {
	src := rand.NewSource(time.Now().Unix())

	return &Rnd{r: rand.New(src)}

}

func (rs *Rnd) RandomTimeoutMs(baseMs int) int {
	div := baseMs / 10

	return baseMs + rs.r.Intn(div) - div/2
}
