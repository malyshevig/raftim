package net

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"raft/src/proto"
	"testing"
)

type NodeMock struct {
	incomingChannel chan proto.MsgEvent
	outgoingChannel chan proto.MsgEvent
}

func (n *NodeMock) GetIncomingChannel() chan proto.MsgEvent {
	return n.incomingChannel
}

func (n *NodeMock) SetIncomingChannel(c chan proto.MsgEvent) {
	n.incomingChannel = c
}

func (n *NodeMock) GetOutgoingChannel() chan proto.MsgEvent {
	return n.outgoingChannel
}

func (n *NodeMock) SetOutgoingChannel(c chan proto.MsgEvent) {
	n.outgoingChannel = c
}

func TestBuildNodesChain(t *testing.T) {

	n_1 := NodeMock{}
	n_2 := NodeMock{}
	n_3 := NodeMock{}

	n := BuildNodesChain(&n_1, &n_2, &n_3)
	if n == nil {
		fmt.Println("Error")
	} else {
		fmt.Println("Ok")
	}

	m := proto.MsgEvent{
		Srcid: 1,
		Dstid: 2,
	}

	n_1.GetOutgoingChannel() <- m
	m1 := <-n_2.GetIncomingChannel()

	assert.True(t, m.Srcid == m1.Srcid)
	assert.True(t, n_1.incomingChannel != nil)
	assert.True(t, n_1.outgoingChannel != nil)
	assert.True(t, n_2.incomingChannel != nil)
	assert.True(t, n_2.outgoingChannel != nil)
	assert.True(t, n_3.incomingChannel != nil)
	assert.True(t, n_3.outgoingChannel != nil)

}
