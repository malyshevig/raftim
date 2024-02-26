package util

import (
	"fmt"
	"github.com/wmentor/latency"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

func Min(a int, b int) int {
	if a < b {
		return a
	}

	return b

}

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	if err == nil {
		err := f.Close()
		if err != nil {
			return false
		}
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
			err := os.Rename(filename, fname)
			if err != nil {
				log.Fatal(err)
				return
			}

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

func GetLogName(id int) string {
	return fmt.Sprintf("./data/nodeLog%d.txt", id)
}

func CallWithMetrix(call func()) time.Duration {
	l := latency.New()
	call()
	return l.Duration()
}
