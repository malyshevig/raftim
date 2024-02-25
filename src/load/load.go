package load

import (
	"fmt"
	"github.com/wmentor/latency"
	"log"
	"raft/src/client"
	"time"
)

func callWithMetrix(call func()) time.Duration {
	l := latency.New()
	call()
	return l.Duration()
}

func Load(node *client.RaftClientNode) {
	log.Printf("Load test start")
	cmdNum := 1
	for c := 0; c < 50000; {
		lat := callWithMetrix(func() {
			err := node.ProcessRequest(fmt.Sprintf("cmd#%d", cmdNum))
			if err != nil {
				return
			}
		})

		fmt.Printf("cmdNum = %d   latency = %d\n", c, lat.Milliseconds())

		//leader.AppendCommand(fmt.Sprintf("cmd#%d", cmdNum))
		cmdNum++
		c++

		//	time.Sleep(time.Millisecond * 10)
	}
	log.Printf("exit Load test")
}
