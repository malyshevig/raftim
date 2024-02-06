package raft

import "time"

func TimeoutTick(tmMs int) {
	for {
		time.Sleep(time.Duration(int64(tmMs) * int64(time.Millisecond)))

		msg := SystemEvent{body: TimerTick{}}

		cluster := ClusterInstance()
		err := cluster.sendAllSysEvent(msg)
		if err != nil {
			panic("Can'not send timer tick event")
		}
	}
}
