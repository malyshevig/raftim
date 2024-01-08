package raft

import "time"

const (
	IdleTimeoutMin = 10
	E
)

func Raft() {
	nw := ClusterInstance()
	nw.NodeAdd(NewNode(1))
	nw.NodeAdd(NewNode(2))
	nw.NodeAdd(NewNode(3))

	for rn := nw.GetNodes().Front(); rn != nil; rn = rn.Next() {
		go rn.Value.(*RaftNode).run()
	}

	go TimeoutTick(500)

	server := RestServer{}
	server.run()
	time.Sleep(1 * time.Hour)
}
