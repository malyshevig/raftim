package raft

import (
	"log"
	"os"
	"raft/src/raftApi"
	"raft/src/util"
)

func (rn *RaftNode) initCmdLog() {
	fname := util.GetLogName(rn.Id)
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
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

func (rn *RaftNode) saveLog(indexFrom int, indexTo int) {

	f, err := os.OpenFile(util.GetLogName(rn.Id), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

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
