package raft

import (
	"time"
)

const (
	CHANNELSIZE = 1000000
)

// IncomingChan: make(chan interface{}, 1000000)
func createNode(id int, r *Router, rbase rnd) *RaftNode {
	routerToNode := make(chan interface{}, CHANNELSIZE)
	delay := CreateDelay(1000, nil, r.incomingChannel)
	node := NewNode(id, rbase, routerToNode, delay.inputChannel)
	r.AddRoute(node.Id, routerToNode)

	go delay.run()

	return node
}

func Raft() {
	rnd := initRand()
	nw := ClusterInstance()

	router := CreateRouter(nil)

	nw.NodeAdd(createNode(1, router, rnd))
	nw.NodeAdd(createNode(2, router, rnd))
	nw.NodeAdd(createNode(3, router, rnd))

	go router.run()

	for rn := nw.GetNodes().Front(); rn != nil; rn = rn.Next() {
		go rn.Value.(*RaftNode).run()
	}

	go TimeoutTick(100)

	server := RestServer{}
	server.run()

	go Load()
	time.Sleep(1 * time.Hour)
}
