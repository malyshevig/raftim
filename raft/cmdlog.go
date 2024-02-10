package raft

import (
	"fmt"
	"log"
	"os"
)

func (rn *RaftNode) checkLog(index int, term int64) bool {
	if index < 0 {
		return true
	}

	if index >= len(rn.CmdLog) {
		return false
	}

	return rn.CmdLog[index].term == term
}

func (rn *RaftNode) appendLogEntry(startIndex int, entry Entry) {
	if startIndex < len(rn.CmdLog) {
		rn.print(fmt.Sprintf("recieved msg index %d already saved len(log)=%d", startIndex, len(rn.CmdLog)))
	} else {
		rn.CmdLog = append(rn.CmdLog, entry)
	}
}

var persist = true

func (rn *RaftNode) saveLog(indexFrom int, indexTo int) {
	if !persist {
		return
	}

	f, err := os.OpenFile(getLogName(rn.id), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
			_, err = f.WriteString(rn.CmdLog[idx].cmd + "\n")
			if err != nil {
				log.Fatal(err)
			}

		}
	}
}
