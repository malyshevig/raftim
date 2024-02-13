package raft

import "container/list"

type Cluster struct {
	Nodes map[int]*RaftNode
}

func (nw *Cluster) NodesCount() int {
	return len(nw.Nodes)
}

func (nw *Cluster) NodeAdd(node *RaftNode) {
	nw.Nodes[node.Node.id] = node
}

func (nw *Cluster) GetNodes() *list.List {
	nodes := list.New()
	for _, n := range nw.Nodes {
		nodes.PushBack(n)
	}
	return nodes
}

func (nw *Cluster) getLeaderId() int {
	var r *RaftNode = nil
	for _, n := range nw.Nodes {
		if n.State == "leader" {
			r = n
		}
	}
	if r != nil {
		return r.id
	} else {
		return 0
	}
}

func (cl *Cluster) getNode(id int) *RaftNode {
	for _, rn := range cl.Nodes {
		if rn.id == id {
			return rn
		}
	}
	return nil
}

func ClusterInstance() *Cluster {
	return &network
}

var network = Cluster{Nodes: make(map[int]*RaftNode)}
