package main

import (
	"go.uber.org/zap"
	"log"
	_ "raft/docs"
	"raft/server"
	"time"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {

		}
	}(logger)

	sugar.Info("Raft Started")

	server.StartServer()

	time.Sleep(1 * time.Hour)
}
