package raft

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func ZeroTime() time.Time {
	return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local)

}

func IsTimeout(ts time.Time, now time.Time, intervalMS int) bool {
	targetTs := ts.Add(time.Duration(intervalMS) * time.Millisecond)

	return now.After(targetTs)
}

func GetLeader() *RaftNode {
	for n := ClusterInstance().GetNodes().Front(); n != nil; n = n.Next() {
		if n.Value.(*RaftNode).State == Leader {
			return n.Value.(*RaftNode)
		}
	}
	return nil
}

type rnd struct {
	r *rand.Rand
}

func initRand() rnd {
	return rnd{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (rs *rnd) RandomTimeoutMs(baseMs int) int {
	div := 4000

	return baseMs + rs.r.Intn(div) - div/2
}

func getLogName(id int) string {
	return fmt.Sprintf("./nodeLog%d.txt", id)
}

func (rn *RaftNode) print(s string) {
	fmt.Printf("node %d term=%d state = %s %s \n", rn.Id, rn.CurrentTerm, rn.State, s)
}

func initTrace() {
	f, err := os.OpenFile("./trace.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
}

func trace(ev MsgEvent) {
	f, err := os.OpenFile("./trace.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer f.Close()

	s := fmt.Sprintf("%v: %d->%d: %s\n", ev.ts, ev.srcid, ev.dstid, ev.body)
	f.WriteString(s)
}
