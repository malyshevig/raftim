package raft

import "time"

func Timer(tmSecs int64, event interface{}, ch chan interface{}) {
	time.Sleep(time.Duration(tmSecs * int64(time.Second)))

	ch <- event
}
