package raft

import "container/list"

type Cluster struct {
	Nodes map[int]*RaftNode
}

func (nw *Cluster) sendAllSysEvent(msg SystemEvent) error {
	for _, rn := range nw.Nodes {
		rn.Node.incomingChan <- msg
	}
	return nil
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
