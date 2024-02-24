package main

import (
	"go.uber.org/zap"
	"log"
	_ "raft/docs"
	"raft/src/client"
	"raft/src/load"
	"raft/src/mgmt"
	nw2 "raft/src/nw"
	"raft/src/raft"
	"raft/src/raftApi"
	"raft/src/rest"
	"time"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {

		}
	}(logger)

	sugar.Info("Raft Started")

	StartServer()

}

//// incomingChan: make(chan interface{}, 1000000)
//func CreateNode(id int, r *nw2.Router, config nw2.ClusterConfig) *raft.RaftNode {
//	routerToNode := make(chan raftApi.MsgEvent, nw2.CHANNELSIZE)
//	delay := nw2.CreateDelay(nw2.DELAY, nil, r.IncomingChannel)
//
//	tm := nw2.RandomiseTimeout(raft.CandidateElectionTimeout)
//	timeouts := raft.Timeout{ElectionTimeoutMS: tm,
//		FollowerTimeoutMS: tm}
//
//	loggerFile := fmt.Sprintf("logs/node%d.log", id)
//
//	node := raft.NewNode(id, timeouts, config, routerToNode, delay.InputChannel, util.InitLogger(loggerFile))
//	r.AddRoute(node.Id, routerToNode)
//
//	go delay.Run()
//
//	return node
//}
//
//func CreateClientNode(id int, r *nw2.Router, config nw2.ClusterConfig) *client.RaftClientNode {
//	routerToNode := make(chan raftApi.MsgEvent, nw2.CHANNELSIZE)
//	delay := nw2.CreateDelay(nw2.DELAY, nil, r.IncomingChannel)
//	node := client.NewClientNode(id, config, routerToNode, delay.InputChannel)
//	r.AddRoute(node.Id, routerToNode)
//
//	go delay.Run()
//	return node
//}

func StartServer() {
	nodesNum := 3
	config := makeClusterConfig(nodesNum)
	cluster := mgmt.ClusterInstance()

	router := nw2.CreateRouter(nil)
	tickGenerator := nw2.NewTickGenerator(make([]chan raftApi.SystemEvent, 0))

	for c := 0; c < nodesNum; c++ {
		node := raft.NewNode(c+1, config)
		delay := nw2.CreateDelay(nw2.DELAY)

		nw2.BuildNodesChain(node, delay, router)

		router.AddRoute(node.Id, node.IncomingChan)
		tickGenerator.Register(node.ControlChan)
		cluster.NodeAdd(node)

		go node.Run()
		go delay.Run()
	}

	clientNode := client.NewClientNode(100, config)
	delay := nw2.CreateDelay(nw2.DELAY)
	nw2.BuildNodesChain(clientNode, delay, router)

	go clientNode.Run()
	go delay.Run()

	router.AddRoute(clientNode.Id, clientNode.IncomingChannel)
	tickGenerator.Register(clientNode.ControlChannel)

	go router.Run()
	go tickGenerator.Run(20)

	server := rest.NewRestServer(clientNode)
	server.Run()

	go load.Load(clientNode)
	//go load.Load(clientNode)
	time.Sleep(1 * time.Hour)
}

func makeClusterConfig(nodesNum int) nw2.ClusterConfig {
	var nodes []int

	for id := 0; id < nodesNum; id++ {
		nodes = append(nodes, id+1)
	}
	config := nw2.ClusterConfig{Nodes: nodes}
	return config
}

//func Raft() {
//	rnd := nw2.InitRand()
//	cluster := mgmt.ClusterInstance()
//
//	router := nw2.CreateRouter(nil)
//
//	config := nw2.ClusterConfig{Nodes: []int{1, 2, 3}}
//	clientNode := CreateClientNode(100, router, config)
//
//	n1 := CreateNode(1, router, config, rnd)
//	cluster.NodeAdd(n1)
//
//	n2 := CreateNode(2, router, config, rnd)
//	cluster.NodeAdd(n2)
//
//	n3 := CreateNode(3, router, config, rnd)
//	cluster.NodeAdd(n3)
//
//	t := nw2.NewTickGenerator(make([]chan raftApi.SystemEvent, 0))
//	t.Register(clientNode.ControlChannel)
//	t.Register(n1.ControlChan)
//	t.Register(n2.ControlChan)
//	t.Register(n3.ControlChan)
//
//	go router.Run()
//	go n1.Run()
//	go n2.Run()
//	go n3.Run()
//	go clientNode.Run()
//
//	go t.Run(20)
//
//	n_1 := CreateNode(4, router, config, rnd)
//	n_2 := CreateNode(5, router, config, rnd)
//	n_3 := CreateNode(6, router, config, rnd)
//
//	var n_1_int nw2.NodeChainInf = n_1
//	var n_2_int nw2.NodeChainInf = n_2
//	var n_3_int nw2.NodeChainInf = n_3
//
//	n := nw2.BuildNodesChain(n_1_int, n_2_int, n_3_int)
//	if n == nil {
//		fmt.Println("Error")
//	} else {
//		fmt.Println("Ok")
//	}
//
//	server := rest.NewRestServer(clientNode)
//	server.Run()
//
//	go load.Load(clientNode)
//	//go load.Load(clientNode)
//	time.Sleep(1 * time.Hour)
//}
