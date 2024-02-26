package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"raft/server"
	"raft/src/util"
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

	for c := 1; c < 10000; c++ {
		lat := util.CallWithMetrix(
			func() {
				cmd := fmt.Sprintf("test_cmd#%d", c)
				sendCommand(client, fmt.Sprintf(cmd))
			})

		fmt.Printf("letency = %d\n", lat/time.Millisecond)
	}

}
