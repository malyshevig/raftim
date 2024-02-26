package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"raft/src/client"
	"raft/src/mgmt"
	"raft/src/raft"
	"sort"
	"strconv"
)

type RestServer struct {
	clientNode *client.RaftClientNode
}

func NewRestServer(clientNode *client.RaftClientNode) *RestServer {
	return &RestServer{clientNode: clientNode}
}

type RestNode struct {
	ID              int    `json:"Id"`
	TERM            int64  `json:"term"`
	STATE           string `json:"state"`
	STATUS          string `json:"status"`
	LEADER          int    `json:"leader"`
	LOG_LEN         int    `json:"log_len"`
	COMMITTED_INDEX int    `json:"comitted_index"`
}

func (r RestNode) String() string {
	res, _ := json.Marshal(r)
	return string(res)
}

// @title           Raft Example
// @version         1.0
// @description     Raft just for fun implementation

// @contact.name   IM
// @contact.email  ilia.malishev@gmail.com

// @host      localhost:1000
// @BasePath  /raft

// Operation     godoc
// @Summary      Get books array
// @Description  Responds with the list of all books as JSON.
// @Tags         nodes
// @Success		200	{array}		raft.RestNode
// @Router       /nodes [get]
func (srv *RestServer) Nodes(c *gin.Context) {
	fmt.Printf("Request GetNodes ")
	restNodes := srv.GetNodes()
	fmt.Printf("GetNodes: %d\n", len(restNodes))
	c.IndentedJSON(http.StatusOK, restNodes)
}

func (srv *RestServer) GetNodes() []RestNode {
	nodes := mgmt.ClusterInstance().GetNodes()
	var restNodes []RestNode
	for n := nodes.Front(); n != nil; n = n.Next() {
		rn := n.Value.(*raft.RaftNode)
		restNodes = append(restNodes, RestNode{ID: rn.Id, TERM: rn.CurrentTerm, STATE: rn.State,
			LEADER: rn.VotedFor, STATUS: rn.Status, COMMITTED_INDEX: rn.CommitedIndex, LOG_LEN: len(rn.CmdLog)})

	}
	sort.Slice(restNodes, func(i, j int) bool {
		return restNodes[i].ID < restNodes[j].ID
	})
	return restNodes
}

// Operation     godoc
// @Summary      Change state of Node
// @Description  Responds with the list of all books as JSON.
// @Tags         node
// @Param			Id	path		int	true	"Node ID"
// @Param			value	path		string	true	"New State"
// @Produce      json
// @Success      200
// @Router       /node/{Id}/state/{value} [post]
func (srv *RestServer) ChangeState(c *gin.Context) {
	fmt.Printf("changeState params= %v\n", c.Params)

	id, err := strconv.Atoi(c.Param("Id"))
	if err != nil {
		msg := fmt.Sprintf("error with parameter Id:%s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": msg})
		return
	}

	state := c.Param("value")

	node, ok := mgmt.ClusterInstance().Nodes[id]
	if !ok {
		r := fmt.Sprintf("node with Id = %d not found", id)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": r})
		return
	}

	err = node.SwitchToState(state)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

// Operation     godoc
// @Summary      Change state of Node
// @Description  Responds with the list of all books as JSON.
// @Tags         node
// @Param			Id	path		int	true	"Node ID"
// @Param			value	path		string	true	"New Status"
// @Produce      json
// @Success      200
// @Router       /node/{Id}/status/{value} [post]
func (srv *RestServer) ChangeStatus(c *gin.Context) {
	fmt.Printf("changeStatus params= %v\n", c.Params)

	id, err := strconv.Atoi(c.Param("Id"))
	if err != nil {
		msg := fmt.Sprintf("error with parameter Id:%s", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": msg})
		return

	}

	status := c.Param("value")

	node, ok := mgmt.ClusterInstance().Nodes[id]
	if !ok {
		r := fmt.Sprintf("node with Id = %d not found", id)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": r})
		return
	}

	node.Status = status
	c.JSON(http.StatusOK, nil)
}

// Operation     godoc
// @Summary      Client Command
// @Description  Responds with the list of all books as JSON.
// @Tags         books
// @Param		 cmd		query		string	false	"string valid"		minlength(1)	maxlength(200)
// @Produce      json
// @Success      200
// @Router       /client/command [post]
func (srv *RestServer) Command(c *gin.Context) {
	cmd := c.Query("cmd")
	err := srv.clientNode.ProcessRequest(cmd)

	if err != nil {
		c.JSON(http.StatusInternalServerError, srv.GetNodes())
	} else {
		c.JSON(http.StatusOK, srv.GetNodes())
	}
	return
}

func (srv *RestServer) Run() {
	gin.SetMode(gin.DebugMode)

	router := gin.Default()
	g1 := router.Group("/raft")

	g1.GET("/nodes", srv.Nodes)
	g1.POST("/node/:Id/state/:value", srv.ChangeState)
	g1.POST("/node/:Id/status/:value", srv.ChangeStatus)
	g2 := router.Group("/raft/client")
	g2.POST("/command", srv.Command)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Println("Routes")
	var cr gin.RouteInfo
	for _, cr = range router.Routes() {
		fmt.Printf("route: %s %s %s %v\n", cr.Path, cr.Method, cr.Handler, cr.HandlerFunc)
	}

	go router.Run(":1000")
}
