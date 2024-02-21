package main

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"raft/client"
	_ "raft/docs"
	"raft/load"
	"raft/mgmt"
	"raft/nw"
	"raft/raft"
	"raft/raftApi"
	"raft/rest"
	"time"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	defer logger.Sync()

	sugar.Info("Raft Started")

	Raft()

}

// incomingChan: make(chan interface{}, 1000000)
func CreateNode(id int, r *nw.Router, config raft.ClusterConfig, rbase *nw.Rnd) *raft.RaftNode {
	routerToNode := make(chan raftApi.MsgEvent, nw.CHANNELSIZE)
	delay := nw.CreateDelay(nw.DELAY, nil, r.IncomingChannel)

	tm := rbase.RandomTimeoutMs(raft.CandidateElectionTimeout)
	timeouts := raft.Timeout{ElectionTimeoutMS: tm,
		FollowerTimeoutMS: tm}

	loggerFile := fmt.Sprintf("logs/node%d.log", id)

	node := raft.NewNode(id, timeouts, config, routerToNode, delay.InputChannel, raft.InitLogger(loggerFile))
	r.AddRoute(node.Id, routerToNode)

	go delay.Run()

	return node
}

func CreateClientNode(id int, r *nw.Router, config raft.ClusterConfig) *client.ClientNode {
	routerToNode := make(chan raftApi.MsgEvent, nw.CHANNELSIZE)
	delay := nw.CreateDelay(nw.DELAY, nil, r.IncomingChannel)
	node := client.NewClientNode(id, config, routerToNode, delay.InputChannel)
	r.AddRoute(node.Id, routerToNode)

	go delay.Run()
	return node
}

func Raft() {
	rnd := nw.InitRand()
	cluster := mgmt.ClusterInstance()

	router := nw.CreateRouter(nil)

	config := raft.ClusterConfig{Nodes: []int{1, 2, 3}}
	clientNode := CreateClientNode(100, router, config)

	n1 := CreateNode(1, router, config, rnd)
	cluster.NodeAdd(n1)

	n2 := CreateNode(2, router, config, rnd)
	cluster.NodeAdd(n2)

	n3 := CreateNode(3, router, config, rnd)
	cluster.NodeAdd(n3)

	t := nw.NewTickGenerator(make([]chan raftApi.SystemEvent, 0))
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

	server := rest.NewRestServer(clientNode)
	server.Run()

	go load.Load(clientNode)
	//go load.Load(clientNode)
	time.Sleep(1 * time.Hour)
}
