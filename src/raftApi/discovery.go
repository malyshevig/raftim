package raftApi

import "fmt"

type LeaderDiscoveryRequest struct {
}

func (v LeaderDiscoveryRequest) String() string {
	return fmt.Sprintf("leader_req()")
}

type LeaderDiscoveryResponse struct {
	LeaderId int
}

func (v LeaderDiscoveryResponse) String() string {
	return fmt.Sprintf("leader_resp(LeaderId: %d)", v.LeaderId)
}
