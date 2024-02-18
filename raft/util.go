package raft

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	if err == nil {
		f.Close()
		return true
	} else {
		return false
	}
}

func rotateLogFile(filename string) {
	if !fileExists(filename) {
		return
	}

	for c := 0; ; c++ {
		fname := fmt.Sprintf("%s_%d", filename, c)
		if !fileExists(fname) {
			os.Rename(filename, fname)
			return
		}
	}
}

func InitLogger(filename string) *zap.Logger {
	rotateLogFile(filename)

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.NameKey = "name"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"

	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{filename}

	return zap.Must(cfg.Build())
}

func getLogName(id int) string {
	return fmt.Sprintf("./nodeLog%d.txt", id)
}

func (rn *RaftNode) print2(s string) {

	fname := fmt.Sprintf("./log_%d.txt", rn.Id)

	f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	str := fmt.Sprintf("%v: node %d term=%d state = %s %s \n", time.Now(), rn.Id, rn.CurrentTerm, rn.State, s)
	_, err = f.WriteString(str)
	if err != nil {
		log.Fatal(err)
	}
}
