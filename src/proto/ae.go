package proto

import "fmt"

type AppendEntries struct {
	Id      int
	Term    int64
	Entries []Entry

	LastLogIndex int
	LastLogTerm  int64

	LeaderCommittedIndex int
}

func (v AppendEntries) String() string {
	s := ""
	for _, e := range v.Entries {
		s = s + e.Cmd + ","
	}

	s = "[" + s + "]"

	return fmt.Sprintf("ae(Id:%d, Term:%d, Entries:%s, laslogIndex: %d, lastlogTerm: %d, LeaderCommittedIndex: %d)",
		v.Id, v.Term, s, v.LastLogIndex, v.LastLogTerm, v.LeaderCommittedIndex)
}

type AppendEntriesResponse struct {
	AeId      int
	Success   bool
	LastIndex int
}

func (v AppendEntriesResponse) String() string {
	return fmt.Sprintf("ae_resp(ae_id: %d, Success:%v, lastIndex:%d)", v.AeId, v.Success, v.LastIndex)
}
