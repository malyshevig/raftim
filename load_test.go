package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"raft/server"
	"raft/src/util"
	"sort"
	"testing"
	"time"
)

func sendCommand(client *http.Client, cmd string) bool {
	var url = "http://localhost:1000/raft/client/command"

	s := fmt.Sprintf("{ 'cmd': %s }", cmd)

	body := []byte(s)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatal(err)
	}
	resp, err2 := client.Do(req)
	if err2 != nil {
		log.Fatal(err)
	}

	return resp.StatusCode == 200
}

func TestLoad(t *testing.T) {
	server.StartServer()
	var client = &http.Client{}
	time.Sleep(time.Second * 10)

	const count = 1000
	results := make([]int, count)

	for c := 0; c < count; c++ {
		lat := util.CallWithMetrix(
			func() {
				cmd := fmt.Sprintf("test_cmd#%d", c)
				sendCommand(client, fmt.Sprintf(cmd))
			})

		results[c] = int(lat / time.Millisecond)
	}
	sort.Ints(results)
	p := 0.9

	pmin := results[0]
	pmax := results[count-1]

	p90 := results[int(math.Round(count*p))-1]
	fmt.Printf("latency pmin = %d, pmax = %d,  p90 = %d\n", pmin, pmax, p90)
}
