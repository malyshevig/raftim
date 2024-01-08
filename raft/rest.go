package raft

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type RestServer struct {
}

type RestNode struct {
	ID     int    `json:"id"`
	TERM   int64  `json:"term"`
	STATE  string `json:"state"`
	LEADER int    `json:"leader"`
}

func (server *RestServer) Nodes(c *gin.Context) {
	nodes := ClusterInstance().GetNodes()
	var restNodes []RestNode
	for n := nodes.Front(); n != nil; n = n.Next() {
		rn := n.Value.(*RaftNode)
		restNodes = append(restNodes, RestNode{ID: rn.Id, TERM: rn.CurrentTerm, STATE: rn.State, LEADER: rn.VotedFor})
		fmt.Printf("%v", rn)
	}

	fmt.Printf("GetNodes: %d\n", len(restNodes))
	c.IndentedJSON(http.StatusOK, restNodes)
}

func (server *RestServer) ChangeState(c *gin.Context) {
	fmt.Printf("changeState params= %v\n", c.Params)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		msg := fmt.Sprintf("error with parameter id:%s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": msg})
		return
	}

	state := c.Param("state")

	node, ok := ClusterInstance().Nodes[id]
	if !ok {
		r := fmt.Sprintf("node with id = %d not found", id)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": r})
		return
	}

	err = node.switchToState(state)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (server *RestServer) run() {
	router := gin.Default()

	router.GET("/nodes", server.Nodes)
	router.POST("/nodes/:id/:state", server.ChangeState)
	go router.Run(":1000")
}
