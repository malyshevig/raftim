package raft

import (
	"fmt"
	"github.com/wmentor/latency"
	"time"
)

func callWithMetrix(call func()) time.Duration {
	l := latency.New()
	call()
	return l.Duration()
}

func Load(node *ClientNode) {
	cmdNum := 1
	for c := 0; c < 5000; {
		lat := callWithMetrix(func() {
			//node.processRequest(fmt.Sprintf("cmd#%d", cmdNum))
		})

		fmt.Printf("cmdNum = %d   latency = %d\n", c, lat.Milliseconds())

		//leader.AppendCommand(fmt.Sprintf("cmd#%d", cmdNum))
		cmdNum++
		c++
		fmt.Printf("cmdNum = %d\n", c)

		//	time.Sleep(time.Millisecond * 10)
	}
	print("exit loader")
}
