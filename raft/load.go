package raft

import (
	"fmt"
)

func Load(node *ClientNode) {
	cmdNum := 1
	for c := 0; c < 5000; {
		leader := GetLeader()
		if leader != nil {
			node.processRequest(fmt.Sprintf("cmd#%d", cmdNum))

			//leader.AppendCommand(fmt.Sprintf("cmd#%d", cmdNum))
			cmdNum++
			c++
			fmt.Printf("cmdNum = %d\n", c)
		}

		//	time.Sleep(time.Millisecond * 10)
	}
	print("exit loader")
}
