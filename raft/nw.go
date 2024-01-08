package raft

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Cluster struct {
	Nodes map[int]*RaftNode
}

func (nw *Cluster) NodesCount() int {
	return len(nw.Nodes)
}

func (nw *Cluster) NodeAdd(node *RaftNode) {
	nw.Nodes[node.Node.Id] = node
}

func (nw *Cluster) GetNodes() *list.List {
	nodes := list.New()
	for _, n := range nw.Nodes {
		nodes.PushBack(n)
	}
	return nodes
}

func (nw *Cluster) send(srcId, dstId int, msg Event) error {
	if srcId > 0 {
		fmt.Printf("Network: send from %d to %d msg=%v \n", srcId, dstId, reflect.TypeOf(msg.Inner))
	}
	rn, ok := nw.Nodes[dstId]
	if !ok {
		return errors.New("not found")
	}

	ch := rn.Node.IncommingChan
	ch <- msg
	return nil
}

func (nw *Cluster) sendAll(srcId int, msg Event) error {
	for _, rn := range nw.Nodes {
		if rn.Node.Id != srcId {
			nw.send(srcId, rn.Id, msg)
		}
	}

	return nil
}

var network Cluster = Cluster{Nodes: make(map[int]*RaftNode)}

func ClusterInstance() *Cluster {
	return &network
}

func TimeoutTick(tmMs int) {
	for {
		time.Sleep(time.Duration(int64(tmMs) * int64(time.Millisecond)))

		msg := Event{Term: 0, Inner: TimerTick{}}
		err := ClusterInstance().sendAll(0, msg)
		if err != nil {
			panic("Can'not send timer tick event")
		}
	}
}
