package raft

import (
	"time"
)

const (
	CHANNELSIZE = 1000000
	DELAY       = 100
)

// incomingChan: make(chan interface{}, 1000000)
func createNode(id int, r *Router, rbase rnd) *RaftNode {
	routerToNode := make(chan MsgEvent, CHANNELSIZE)
	delay := CreateDelay(DELAY, nil, r.incomingChannel)
	node := NewNode(id, rbase, routerToNode, delay.inputChannel)
	r.AddRoute(node.id, routerToNode)

	go delay.run()

	return node
}

func createClientNode(id int, r *Router, rbase rnd) *ClientNode {
	routerToNode := make(chan MsgEvent, CHANNELSIZE)
	delay := CreateDelay(DELAY, nil, r.incomingChannel)
	node := NewClientNode(id, routerToNode, delay.inputChannel)
	r.AddRoute(node.id, routerToNode)

	go delay.run()
	return node
}

func Raft() {
	rnd := initRand()
	nw := ClusterInstance()

	router := CreateRouter(nil)

	clientNode := createClientNode(100, router, rnd)

	n1 := createNode(1, router, rnd)
	nw.NodeAdd(n1)

	n2 := createNode(2, router, rnd)
	nw.NodeAdd(n2)

	n3 := createNode(3, router, rnd)
	nw.NodeAdd(n3)

	t := TickGenerator{channels: make([]chan SystemEvent, 0)}
	t.addChan(clientNode.controlChannel)
	t.addChan(n1.controlChan)
	t.addChan(n2.controlChan)
	t.addChan(n3.controlChan)

	clientNode.addNodes(n1.id, n2.id, n3.id)

	go router.run()
	go n1.run()
	go n2.run()
	go n3.run()
	go clientNode.run()

	go t.run(100)

	server := RestServer{clientNode: clientNode}
	server.run()

	go Load(clientNode)
	time.Sleep(1 * time.Hour)
}
