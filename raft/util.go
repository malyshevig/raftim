package raft

import (
	"math/rand"
	"time"
)

func Timer(tmSecs int64, event Event, ch chan Event) {
	time.Sleep(time.Duration(tmSecs * int64(time.Second)))

	ch <- event
}

func RandomTimeout(min, max int) int {
	return min + rand.Intn(max-min)
}

func ZeroTime() time.Time {
	return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local)

}

func IsTimeout(ts time.Time, now time.Time, intervalSec int) bool {
	targetTs := ts.Add(time.Duration(time.Duration(intervalSec) * time.Second))

	return now.After(targetTs)
}
