package main

import (
	"bytes"
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

func getNodes(client *http.Client) []rest.RestNode {
	url := "http://localhost:1000/raft/nodes"

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
	var client = &http.Client{}

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
		if c.STATE == "leader" && c.STATUS == "active" {
			leader = c
		}
	}
	return &leader
}

func checkNodes(nodes []rest.RestNode, leaderId int) bool {
	for _, c := range nodes {
		if c.STATE == "active" && c.LEADER != leaderId {

			return false
		}
	}
	return true
}

func testKillLeader(t *testing.T, client *http.Client) {

	nodes := getNodes(client)
	leader := getLeader(nodes)

	assert.True(t, leader.ID > 0)
	assert.True(t, checkNodes(nodes, leader.ID))
	fmt.Printf("result = %v\n", leader)

	killLeader(leader)
	time.Sleep(time.Second * 10)

	nodes = getNodes(client)
	leader = getLeader(nodes)
	fmt.Printf("newLeader = %v\n", leader)
	assert.True(t, leader.ID > 0)
	assert.True(t, checkNodes(nodes, leader.ID))
}

func testPostCommand(t *testing.T, client *http.Client) {
	url := fmt.Sprintf("http://localhost:1000/raft/client/command")

	body := []byte(`{
		"cmd": "Hello",
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatal(err)
	}
	resp, err2 := client.Do(req)
	if err2 != nil {
		log.Fatal(err)
	}
	assert.True(t, resp.StatusCode == 200)
}

func TestPack(t *testing.T) {
	server.StartServer()
	var client = &http.Client{}
	time.Sleep(time.Second * 10)

	t.Run("killLeader", func(t *testing.T) {
		testKillLeader(t, client)
	})
	t.Run("postCommands", func(t *testing.T) {
		testPostCommand(t, client)
	})
}
