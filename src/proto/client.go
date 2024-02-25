package proto

import "fmt"

type ClientCommand struct {
	Id  int64
	Cmd string
}

func (v ClientCommand) String() string {
	return fmt.Sprintf("Cmd(Id: %d, Cmd:%s )", v.Id, v.Cmd)
}

type ClientCommandResponse struct {
	CmdId   int64
	Success bool

	Leaderid int
}

func (v ClientCommandResponse) String() string {
	return fmt.Sprintf("cmd_resp(Id: %d, Success:%v Leaderid:%d )", v.CmdId, v.Success, v.Leaderid)
}

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
