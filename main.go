package main

import (
	"fmt"
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
	"raft/src/util"
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

	Raft()

}

// incomingChan: make(chan interface{}, 1000000)
func CreateNode(id int, r *nw2.Router, config raftApi.ClusterConfig, rbase *nw2.Rnd) *raft.RaftNode {
	routerToNode := make(chan raftApi.MsgEvent, nw2.CHANNELSIZE)
	delay := nw2.CreateDelay(nw2.DELAY, nil, r.IncomingChannel)

	tm := rbase.RandomiseTimeout(raft.CandidateElectionTimeout)
	timeouts := raft.Timeout{ElectionTimeoutMS: tm,
		FollowerTimeoutMS: tm}

	loggerFile := fmt.Sprintf("logs/node%d.log", id)

	node := raft.NewNode(id, timeouts, config, routerToNode, delay.InputChannel, util.InitLogger(loggerFile))
	r.AddRoute(node.Id, routerToNode)

	go delay.Run()

	return node
}

func CreateClientNode(id int, r *nw2.Router, config raftApi.ClusterConfig) *client.ClientNode {
	routerToNode := make(chan raftApi.MsgEvent, nw2.CHANNELSIZE)
	delay := nw2.CreateDelay(nw2.DELAY, nil, r.IncomingChannel)
	node := client.NewClientNode(id, config, routerToNode, delay.InputChannel)
	r.AddRoute(node.Id, routerToNode)

	go delay.Run()
	return node
}

func Raft() {
	rnd := nw2.InitRand()
	cluster := mgmt.ClusterInstance()

	router := nw2.CreateRouter(nil)

	config := raftApi.ClusterConfig{Nodes: []int{1, 2, 3}}
	clientNode := CreateClientNode(100, router, config)

	n1 := CreateNode(1, router, config, rnd)
	cluster.NodeAdd(n1)

	n2 := CreateNode(2, router, config, rnd)
	cluster.NodeAdd(n2)

	n3 := CreateNode(3, router, config, rnd)
	cluster.NodeAdd(n3)

	t := nw2.NewTickGenerator(make([]chan raftApi.SystemEvent, 0))
	t.AddChan(clientNode.ControlChannel)
	t.AddChan(n1.ControlChan)
	t.AddChan(n2.ControlChan)
	t.AddChan(n3.ControlChan)

	go router.Run()
	go n1.Run()
	go n2.Run()
	go n3.Run()
	go clientNode.Run()

	go t.Run(20)

	n_1 := CreateNode(4, router, config, rnd)
	n_2 := CreateNode(5, router, config, rnd)
	n_3 := CreateNode(6, router, config, rnd)

	var n_1_int nw2.NodeChainInf = n_1
	var n_2_int nw2.NodeChainInf = n_2
	var n_3_int nw2.NodeChainInf = n_3

	n := nw2.BuildNodesChain(n_1_int, n_2_int, n_3_int)
	if n == nil {
		fmt.Println("Error")
	} else {
		fmt.Println("Ok")
	}

	server := rest.NewRestServer(clientNode)
	server.Run()

	go load.Load(clientNode)
	//go load.Load(clientNode)
	time.Sleep(1 * time.Hour)
}
