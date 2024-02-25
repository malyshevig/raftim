package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type CmdProcess struct {
	wg *sync.WaitGroup

	id    int64
	cmd   string
	state string
}

func (c CmdProcess) String() string {
	return fmt.Sprintf("%s", c.cmd)
}

func (cl *RaftClientNode) processClientCmd(cmd *CmdProcess) {
	cl.cmdList.PushBack(cmd)
	cl.sendCommands()
}

func (cl *RaftClientNode) commitCommands(cmdid int64) {
	cl.logger.Infof("%s commitCommands %d", *cl, cmdid)
	fmt.Printf("%v Client commitCommands %d\n", time.Now(), cmdid)
	for c := cl.cmdSentList.Front(); c != nil; c = c.Next() {
		cmd := c.Value.(*CmdProcess)
		cmd.state = "Done"
		cmd.wg.Done()

		cl.cmdSentList.Remove(c)
		if cmd.id == cmdid {
			break
		}
	}
}

func (cl *RaftClientNode) allocateCmdId() int64 {
	for {
		v := cl.clientCommandId
		if atomic.CompareAndSwapInt64(&cl.clientCommandId, v, v+1) {
			return v + 1
		}
	}
}
