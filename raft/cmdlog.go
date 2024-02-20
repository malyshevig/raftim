package raft

import (
	"log"
	"os"
	"raft/raftApi"
)

func (rn *RaftNode) initCmdLog() {
	fname := getLogName(rn.Id)
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
}

func (rn *RaftNode) checkLog(index int, term int64) bool {
	if index < 0 {
		return true
	}

	if index >= len(rn.CmdLog) {
		return false
	}

	return rn.CmdLog[index].Term == term
}

func (rn *RaftNode) appendLogEntry(startIndex int, entry raftApi.Entry) {
	if startIndex < len(rn.CmdLog) {
		rn.logger.Infof("%s recieved msg index %d already saved len(log)=%d", *rn, startIndex, len(rn.CmdLog))
	} else {
		rn.CmdLog = append(rn.CmdLog, entry)
	}
}

var persist = true

func (rn *RaftNode) saveLog(indexFrom int, indexTo int) {
	if !persist {
		return
	}

	f, err := os.OpenFile(getLogName(rn.Id), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer f.Close()

	if indexTo >= len(rn.CmdLog) {
		indexTo = len(rn.CmdLog) - 1
	}

	if indexFrom <= indexTo {
		for idx := indexFrom; idx <= indexTo; idx++ {
			_, err = f.WriteString(rn.CmdLog[idx].Cmd + "\n")
			if err != nil {
				log.Fatal(err)
			}

		}
	}
}
