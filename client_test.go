package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"raft/server"
	"raft/src/rest"
	"testing"
	"time"
)

var url = "http://localhost:1000/raft/nodes"
var client = &http.Client{}

func getNodes() []rest.RestNode {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {

	}

	var nodes []rest.RestNode

	err = json.NewDecoder(resp.Body).Decode(&nodes)
	return nodes
}

func killLeader(node *rest.RestNode) bool {
	url := fmt.Sprintf("http://localhost:1000/raft/node/%d/status/fail", node.ID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err2 := client.Do(req)
	if err2 != nil {
		log.Fatal(err)
	}
	return resp.StatusCode == 200
}

func getLeader(nodes []rest.RestNode) *rest.RestNode {
	var leader rest.RestNode
	for _, c := range nodes {
		if c.STATE == "leader" {
			leader = c
		}
	}
	return &leader
}

func checkNodes(nodes []rest.RestNode, leaderId int) bool {
	for _, c := range nodes {
		if c.LEADER != leaderId {
			return false
		}
	}
	return true
}

func TestCheckLeader(t *testing.T) {
	server.StartServer()

	time.Sleep(time.Second * 10)

	nodes := getNodes()
	leader := getLeader(nodes)

	assert.True(t, leader.ID > 0)
	assert.True(t, checkNodes(nodes, leader.ID))
	fmt.Printf("result = %v\n", leader)

	killLeader(leader)

	nodes = getNodes()
	leader = getLeader(nodes)
	fmt.Printf("newLeader = %v\n", leader)
	assert.True(t, leader.ID > 0)
	assert.True(t, checkNodes(nodes, leader.ID))

}

/*
func TestCheckKillLeader(t *testing.T) {
	server.StartServer()

	time.Sleep(time.Second * 10)

	nodes := getNodes()
	leader := getLeader(nodes)

	assert.True(t, leader.ID > 0)
	assert.True(t, checkNodes(nodes, leader.ID))
	fmt.Printf("result = %v\n", leader)
}
*/
