package server

import (
	"raft/src/client"
	"raft/src/mgmt"
	"raft/src/net"
	"raft/src/proto"
	"raft/src/raft"
	"raft/src/rest"
)

func StartServer() {
	nodesNum := 33
	config := makeClusterConfig(nodesNum)
	cluster := mgmt.ClusterInstance()

	router := net.CreateRouter(nil)
	tickGenerator := net.NewTickGenerator(make([]chan proto.SystemEvent, 0))

	for c := 0; c < nodesNum; c++ {
		node := raft.NewNode(c+1, config)
		delay := net.CreateDelay(net.DELAY)

		net.BuildNodesChain(node, delay, router)

		router.AddRoute(node.Id, node.IncomingChan)
		tickGenerator.Register(node.ControlChan)
		cluster.NodeAdd(node)

		go node.Run()
		go delay.Run()
	}

	clientNode := client.NewClientNode(100, config)
	delay := net.CreateDelay(net.DELAY)
	net.BuildNodesChain(clientNode, delay, router)

	go clientNode.Run()
	go delay.Run()

	router.AddRoute(clientNode.Id, clientNode.IncomingChannel)
	tickGenerator.Register(clientNode.ControlChannel)

	go router.Run()
	go tickGenerator.Run(20)

	server := rest.NewRestServer(clientNode)
	server.Run()

	//go load.Load(clientNode)
	//go load.Load(clientNode)
}

func makeClusterConfig(nodesNum int) net.ClusterConfig {
	var nodes []int

	for id := 0; id < nodesNum; id++ {
		nodes = append(nodes, id+1)
	}
	config := net.ClusterConfig{Nodes: nodes}
	return config
}
