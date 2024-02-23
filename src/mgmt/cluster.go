package mgmt

import (
	"container/list"
	"raft/src/raft"
)

type Cluster struct {
	Nodes map[int]*raft.RaftNode
}

func (nw *Cluster) NodesCount() int {
	return len(nw.Nodes)
}

func (nw *Cluster) NodeAdd(node *raft.RaftNode) {
	nw.Nodes[node.Node.Id] = node
}

func (nw *Cluster) GetNodes() *list.List {
	nodes := list.New()
	for _, n := range nw.Nodes {
		nodes.PushBack(n)
	}
	return nodes
}

func (nw *Cluster) getLeaderId() int {
	var r *raft.RaftNode = nil
	for _, n := range nw.Nodes {
		if n.State == "leader" {
			r = n
		}
	}
	if r != nil {
		return r.Id
	} else {
		return 0
	}
}

func ClusterInstance() *Cluster {
	return &network
}

var network = Cluster{Nodes: make(map[int]*raft.RaftNode)}
