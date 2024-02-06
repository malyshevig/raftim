package raft

import (
	"fmt"
	"time"
)

func Load() {
	cmdNum := 1
	for c := 0; c < 5000; {
		leader := GetLeader()
		if leader != nil {
			leader.AppendCommand(fmt.Sprintf("cmd#%d", cmdNum))
			cmdNum++
			c++
		}

		time.Sleep(time.Millisecond * 10)
	}
	print("exit loader")
}
